package handler

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/mrjosh/helm-ls/internal/charts"
	lspinternal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"

	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/mrjosh/helm-ls/pkg/chart"
	"github.com/mrjosh/helm-ls/pkg/chartutil"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (h *langHandler) Hover(ctx context.Context, params *lsp.HoverParams) (result *lsp.Hover, err error) {
	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return nil, errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	chart, err := h.chartStore.GetChartForDoc(params.TextDocument.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", params.TextDocument.URI, err)
	}

	var (
		currentNode = lspinternal.NodeAtPosition(doc.Ast, params.Position)
		parent      = currentNode.Parent()
		wordRange   = lspinternal.GetLspRangeForNode(currentNode)
		word        string
	)

	if parent == nil {
		err = errors.New("Could not parse ast correctly")
		return nil, err
	}

	pt := parent.Type()
	ct := currentNode.Type()
	if ct == gotemplate.NodeTypeText {
		word := doc.WordAt(params.Position)
		if len(word) > 2 && string(word[len(word)-1]) == ":" {
			word = word[0 : len(word)-1]
		}
		response, err := h.yamllsConnector.CallHover(ctx, *params, word)
		return response, err
	}
	if pt == gotemplate.NodeTypeFunctionCall && ct == gotemplate.NodeTypeIdentifier {
		word = currentNode.Content([]byte(doc.Content))
	}
	if (pt == gotemplate.NodeTypeSelectorExpression || pt == gotemplate.NodeTypeField) &&
		(ct == gotemplate.NodeTypeIdentifier || ct == gotemplate.NodeTypeFieldIdentifier) {
		word = lspinternal.GetFieldIdentifierPath(currentNode, doc)
	}
	if ct == gotemplate.NodeTypeDot {
		word = lspinternal.TraverseIdentifierPathUp(currentNode, doc)
	}

	var (
		splitted         = strings.Split(word, ".")
		variableSplitted = []string{}
		value            string
	)

	if word == "" {
		return nil, err
	}

	for _, s := range splitted {
		if s != "" {
			variableSplitted = append(variableSplitted, s)
		}
	}

	// // $ always points to the root context so we must remove it before looking up tables
	if variableSplitted[0] == "$" && len(variableSplitted) > 1 {
		variableSplitted = variableSplitted[1:]
	}

	logger.Println(fmt.Sprintf("Hover checking for word < %s >", word))

	if len(variableSplitted) > 1 {
		switch variableSplitted[0] {
		case "Values":
			value, err = h.getValueHover(chart, variableSplitted[1:])
		case "Chart":
			value, err = h.getChartMetadataHover(&chart.ChartMetadata.Metadata, variableSplitted[1])
		case "Release":
			value, err = h.getBuiltInObjectsHover(releaseVals, variableSplitted[1])
		case "Files":
			value, err = h.getBuiltInObjectsHover(filesVals, variableSplitted[1])
		case "Capabilities":
			value, err = h.getBuiltInObjectsHover(capabilitiesVals, variableSplitted[1])
		}

		if err == nil {
			if value == "" {
				value = "\"\""
			}
			result := util.BuildHoverResponse(value, wordRange)
			return result, err
		}
		return nil, err
	}

	searchWord := variableSplitted[0]
	completionItems := [][]HelmDocumentation{
		basicItems,
		builtinFuncs,
		sprigFuncs,
		helmFuncs,
	}
	toSearch := util.ConcatMultipleSlices(completionItems)

	logger.Println("Start search with word " + searchWord)
	for _, completionItem := range toSearch {
		if searchWord == completionItem.Name {
			result := util.BuildHoverResponse(fmt.Sprint(completionItem.Doc), wordRange)
			return result, err
		}
	}
	return nil, err
}

func (h *langHandler) getChartMetadataHover(metadata *chart.Metadata, key string) (string, error) {
	for _, completionItem := range chartVals {
		if key == completionItem.Name {
			logger.Println("Getting metadatafield of " + key)

			documentation := completionItem.Doc
			value := h.getMetadataField(metadata, key)

			return fmt.Sprintf("%s\n\n%s\n", documentation, value), nil
		}
	}
	return "", fmt.Errorf("%s was no known Chart Metadata property", key)
}

func (h *langHandler) getValueHover(chart *charts.Chart, splittedVar []string) (result string, err error) {
	var (
		valuesFiles = chart.ResolveValueFiles(splittedVar, h.chartStore)
		results     = map[uri.URI]string{}
	)

	for _, valuesFiles := range valuesFiles {
		for _, valuesFile := range valuesFiles.ValuesFiles.AllValuesFiles() {
			result, err := h.getTableOrValueForSelector(valuesFile.Values, strings.Join(valuesFiles.Selector, "."))
			if err == nil {
				results[valuesFile.URI] = result
			}
		}
	}

	keys := make([]string, 0, len(results))
	for u := range results {
		keys = append(keys, string(u))
	}

	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	for _, key := range keys {
		uriKey := uri.New(key)
		value := results[uriKey]
		if value == "" {
			value = "\"\""
		}
		filepath, err := filepath.Rel(h.chartStore.RootURI.Filename(), uriKey.Filename())
		if err != nil {
			filepath = uriKey.Filename()
		}
		result += fmt.Sprintf("### %s\n%s\n\n", filepath, value)
	}
	return result, nil
}

func (h *langHandler) getTableOrValueForSelector(values chartutil.Values, selector string) (string, error) {
	if len(selector) > 0 {
		localValues, err := values.Table(selector)
		if err != nil {
			logger.Debug("values.PathValue(tableName) because of error", err)
			value, err := values.PathValue(selector)
			return h.formatToYAML(reflect.Indirect(reflect.ValueOf(value)), selector), err
		}
		logger.Debug("converting to YAML", localValues)
		return localValues.YAML()
	}
	return values.YAML()
}

func (h *langHandler) getBuiltInObjectsHover(items []HelmDocumentation, key string) (string, error) {
	for _, completionItem := range items {
		if key == completionItem.Name {
			documentation := completionItem.Doc
			return fmt.Sprintf("%s\n", documentation), nil
		}
	}
	return "", fmt.Errorf("%s was no known built-in object", key)
}

func (h *langHandler) getMetadataField(v *chart.Metadata, fieldName string) string {
	r := reflect.ValueOf(v)
	field := reflect.Indirect(r).FieldByName(fieldName)
	return h.formatToYAML(field, fieldName)
}

func (h *langHandler) formatToYAML(field reflect.Value, fieldName string) string {
	switch field.Kind() {
	case reflect.String:
		return field.String()
	case reflect.Map:
		return h.toYAML(field.Interface())
	case reflect.Slice:
		return h.toYAML(map[string]interface{}{fieldName: field.Interface()})
	case reflect.Bool:
		return fmt.Sprint(h.getBoolType(field))
	default:
		return "<Unknown>"
	}
}
