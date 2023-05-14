package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	lspinternal "github.com/mrjosh/helm-ls/internal/lsp"

	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/mrjosh/helm-ls/pkg/chart"
	"github.com/mrjosh/helm-ls/pkg/chartutil"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) handleHover(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {

	if req.Params() == nil {
		return &jsonrpc2.Error{Code: jsonrpc2.InvalidParams}
	}

	var params lsp.HoverParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}

	var (
		currentNode = lspinternal.NodeAtPosition(doc.Ast, params.Position)
		parent      = currentNode.Parent()
		wordRange   = lspinternal.GetLspRangeForNode(currentNode)
		word        string
	)

	if parent == nil {
		err = errors.New("Could not parse ast correctly.")
		return reply(ctx, nil, err)
	}

	pt := parent.Type()
	ct := currentNode.Type()
	if ct == "text" {
		var word = doc.WordAt(params.Position)
		if len(word) > 2 && string(word[len(word)-1]) == ":" {
			word = word[0 : len(word)-1]
		}
		var response = h.yamllsConnector.CallHover(params, word)
		return reply(ctx, response, err)
	}
	if pt == "function_call" && ct == "identifier" {
		word = currentNode.Content([]byte(doc.Content))
	}
	if (pt == "selector_expression" || pt == "field") && (ct == "identifier" || ct == "field_identifier") {
		word = lspinternal.GetFieldIdentifierPath(currentNode, doc)
	}
	if ct == "dot" {
		word = lspinternal.TraverseIdentifierPathUp(currentNode, doc)
	}

	var (
		splitted         = strings.Split(word, ".")
		variableSplitted = []string{}
		value            string
	)

	if word == "" {
		return reply(ctx, nil, err)
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
			value, err = h.getValueHover(variableSplitted[1:])
		case "Chart":
			value, err = h.getChartMetadataHover(variableSplitted[1])
		case "Release":
			value, err = h.getBuiltInObjectsHover(releaseVals, variableSplitted[1])
		case "Files":
			value, err = h.getBuiltInObjectsHover(filesVals, variableSplitted[1])
		case "Capabilities":
			value, err = h.getBuiltInObjectsHover(capabilitiesVals, variableSplitted[1])
		}

		if err == nil {
			result := util.BuildHoverResponse(value, wordRange)
			return reply(ctx, result, err)
		}
		return reply(ctx, nil, err)
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
			return reply(ctx, result, err)
		}
	}
	return reply(ctx, lsp.Hover{}, err)
}

func (h *langHandler) getChartMetadataHover(key string) (string, error) {
	for _, completionItem := range chartVals {
		if key == completionItem.Name {
			logger.Println("Getting metadatafield of " + key)

			documentation := completionItem.Doc
			value := h.getMetadataField(&h.chartMetadata, key)

			return fmt.Sprintf("%s\n\n%s\n", documentation, value), nil
		}
	}
	return "", fmt.Errorf("%s was no known Chart Metadata property", key)
}

func (h *langHandler) getValueHover(splittedVar []string) (string, error) {
	var (
		values      = h.values
		tableName   = strings.Join(splittedVar, ".")
		err         error
		localValues chartutil.Values
		value       interface{}
	)

	if len(splittedVar) > 0 {
		localValues, err = values.Table(tableName)
		if err != nil {
			logger.Println(err)
			logger.Println("values.PathValue(tableName)")
			value, err = values.PathValue(tableName)
			return fmt.Sprint(value), err
		}
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
	switch field.Kind() {
	case reflect.String:
		return field.String()
	case reflect.Slice, reflect.Map:
		return h.toYAML(field.Interface())
	case reflect.Bool:
		return fmt.Sprint(h.getBoolType(field))
	default:
		return "<Unknown>"
	}
}
