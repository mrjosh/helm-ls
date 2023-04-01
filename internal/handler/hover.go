package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

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

	var params lsp.DefinitionParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}

	var (
		word             = doc.ValueAt(params.Position)
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

	// $ always points to the root context so we can safely remove it
	// as long the LSP does not know about ranges
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
		case "Files":
			value, err = h.getFilesHover(variableSplitted[1])
		case "Capabilities":
		default:
		}
	}

	if err == nil {
		content := lsp.MarkupContent{
			Kind:  lsp.Markdown,
			Value: value,
		}
		result := lsp.Hover{
			Contents: content,
			// TODO: could add a range
		}

		return reply(ctx, result, err)
	}
	searchWord := splitted[len(splitted)-1]
	completionItems := [][]lsp.CompletionItem{
		basicItems,
		builtinFuncs,
		sprigFuncs,
		helmFuncs,
		h.getReleaseVals(),
		h.getCapabilitiesVals(),
		h.getFilesVals(),
	}
	toSearch := util.ConcatMultipleSlices(completionItems)

	logger.Println("Start search with word " + searchWord)
	for _, completionItem := range toSearch {
		if searchWord == completionItem.InsertText {

			content := lsp.MarkupContent{
				Kind:  lsp.Markdown,
				Value: fmt.Sprint(completionItem.Documentation),
			}
			result := lsp.Hover{
				Contents: content,
				// TODO: could add a range
			}

			return reply(ctx, result, err)
		}
	}
	return reply(ctx, lsp.Hover{}, err)
}

func (h *langHandler) getChartMetadataHover(key string) (string, error) {
	for _, completionItem := range h.getChartVals() {
		if key == completionItem.InsertText {
			logger.Println("Getting metadatafield of " + key)

			documentation := completionItem.Documentation

			value := h.getMetadataField(&h.chartMetadata, key)

			return fmt.Sprintf("%s\n\n%s\n", documentation, value), nil
		}
	}
	return "", fmt.Errorf("%s was no known Chart Metadata property", key)
}

func (h *langHandler) getValueHover(splittedVar []string) (string, error) {

	var (
		values      = h.values
		err         error
		tableName   = strings.Join(splittedVar, ".")
		localValues chartutil.Values
		value       interface{}
	)

	if len(splittedVar) > 0 {
		localValues, err = values.Table(tableName)
		if err != nil {
			value, err = values.PathValue(tableName)
			logger.Println(err)
			return fmt.Sprint(value), err
		}
		return localValues.YAML()

	}
	return values.YAML()

}

func (h *langHandler) getFilesHover(key string) (string, error) {
	for _, completionItem := range h.getFilesVals() {
		if key == completionItem.InsertText {
			documentation := completionItem.Documentation
			return fmt.Sprintf("%s\n", documentation), nil
		}
	}
	return "", fmt.Errorf("%s was no known Chart Metadata property", key)

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
