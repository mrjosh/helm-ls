package handler

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/mrjosh/helm-ls/internal/charts"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	gotemplate "github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/pkg/chartutil"
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"
	lsp "go.lsp.dev/protocol"
	yaml "gopkg.in/yaml.v2"

	"github.com/mrjosh/helm-ls/internal/documentation/godocs"
)

var (
	emptyItems               = make([]lsp.CompletionItem, 0)
	functionsCompletionItems = make([]lsp.CompletionItem, 0)
	textCompletionsItems     = make([]lsp.CompletionItem, 0)
)

func init() {
	functionsCompletionItems = append(functionsCompletionItems, getFunctionCompletionItems(helmFuncs)...)
	functionsCompletionItems = append(functionsCompletionItems, getFunctionCompletionItems(builtinFuncs)...)
	functionsCompletionItems = append(functionsCompletionItems, getFunctionCompletionItems(sprigFuncs)...)
	textCompletionsItems = append(textCompletionsItems, getTextCompletionItems(godocs.TextSnippets)...)
}

func (h *langHandler) Completion(ctx context.Context, params *lsp.CompletionParams) (result *lsp.CompletionList, err error) {
	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return nil, errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	chart, err := h.chartStore.GetChartForDoc(params.TextDocument.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", params.TextDocument.URI, err)
	}

	word, isTextNode := completionAstParsing(doc, params.Position)

	if isTextNode {
		result := make([]lsp.CompletionItem, 0)
		result = append(result, textCompletionsItems...)
		result = append(result, yamllsCompletions(ctx, h, params)...)
		return &protocol.CompletionList{IsIncomplete: false, Items: result}, err
	}

	var (
		splitted         = strings.Split(word, ".")
		items            []lsp.CompletionItem
		variableSplitted = []string{}
	)

	for n, s := range splitted {
		// we want to keep the last empty string to be able
		// distinguish between 'global.' and 'global'
		if s == "" && n != len(splitted)-1 {
			continue
		}
		variableSplitted = append(variableSplitted, s)
	}

	logger.Println(fmt.Sprintf("Word found for completions is < %s >", word))

	items = make([]lsp.CompletionItem, 0)
	for _, v := range basicItems {
		items = append(items, lsp.CompletionItem{
			Label:         v.Name,
			InsertText:    v.Name,
			Detail:        v.Detail,
			Documentation: v.Doc,
		})
	}
	if len(variableSplitted) == 0 {
		return &lsp.CompletionList{IsIncomplete: false, Items: items}, err
	}

	// $ always points to the root context so we can safely remove it
	// as long the LSP does not know about ranges
	if variableSplitted[0] == "$" && len(variableSplitted) > 1 {
		variableSplitted = variableSplitted[1:]
	}

	switch variableSplitted[0] {
	case "Chart":
		items = getVariableCompletionItems(chartVals)
	case "Values":
		items = h.getValuesCompletions(chart, variableSplitted[1:])
	case "Release":
		items = getVariableCompletionItems(releaseVals)
	case "Files":
		items = getVariableCompletionItems(filesVals)
	case "Capabilities":
		items = getVariableCompletionItems(capabilitiesVals)
	default:
		items = getVariableCompletionItems(basicItems)
		items = append(items, functionsCompletionItems...)
	}

	return &lsp.CompletionList{IsIncomplete: false, Items: items}, err
}

func yamllsCompletions(ctx context.Context, h *langHandler, params *lsp.CompletionParams) []lsp.CompletionItem {
	response, err := h.yamllsConnector.CallCompletion(ctx, params)
	if err != nil {
		logger.Error("Error getting yamlls completions", err)
		return []lsp.CompletionItem{}
	}
	logger.Debug("Got completions from yamlls", response)
	return response.Items
}

func completionAstParsing(doc *lsplocal.Document, position lsp.Position) (string, bool) {
	var (
		currentNode   = lsplocal.NodeAtPosition(doc.Ast, position)
		pointToLoopUp = sitter.Point{
			Row:    position.Line,
			Column: position.Character,
		}
		relevantChildNode = lsplocal.FindRelevantChildNode(currentNode, pointToLoopUp)
		word              string
	)

	logger.Debug("currentNode", currentNode)
	logger.Debug("relevantChildNode", relevantChildNode)

	switch relevantChildNode.Type() {
	case gotemplate.NodeTypeIdentifier:
		word = relevantChildNode.Content([]byte(doc.Content))
	case gotemplate.NodeTypeDot:
		logger.Debug("TraverseIdentifierPathUp for dot node")
		word = lsplocal.TraverseIdentifierPathUp(relevantChildNode, doc)
	case gotemplate.NodeTypeDotSymbol:
		logger.Debug("GetFieldIdentifierPath")
		word = lsplocal.GetFieldIdentifierPath(relevantChildNode, doc)
	case gotemplate.NodeTypeText, gotemplate.NodeTypeTemplate:
		return word, true
	}
	return word, false
}

func (h *langHandler) getValuesCompletions(chart *charts.Chart, splittedVar []string) (result []lsp.CompletionItem) {
	m := make(map[string]lsp.CompletionItem)
	for _, queriedValuesFiles := range chart.ResolveValueFiles(splittedVar, h.chartStore) {
		for _, valuesFile := range queriedValuesFiles.ValuesFiles.AllValuesFiles() {
			for _, item := range h.getValue(valuesFile.Values, queriedValuesFiles.Selector) {
				m[item.InsertText] = item
			}
		}
	}

	for _, item := range m {
		result = append(result, item)
	}

	return result
}

func (h *langHandler) getValue(values chartutil.Values, splittedVar []string) []lsp.CompletionItem {
	var (
		err         error
		tableName   = strings.Join(splittedVar, ".")
		localValues chartutil.Values
		items       = make([]lsp.CompletionItem, 0)
	)

	if len(splittedVar) > 0 {

		localValues, err = values.Table(tableName)
		if err != nil {
			logger.Println(err)
			if len(splittedVar) > 1 {
				// the current tableName was not found, maybe because it is incomplete, we can use the previous one
				// e.g. gobal.im -> im was not found
				// but global contains the key image, so we return all keys of global
				localValues, err = values.Table(strings.Join(splittedVar[:len(splittedVar)-1], "."))
				if err != nil {
					logger.Println(err)
					return emptyItems
				}
				values = localValues
			}
		} else {
			values = localValues
		}

	}

	for variable, value := range values {
		items = h.setItem(items, value, variable)
	}

	return items
}

func (h *langHandler) setItem(items []lsp.CompletionItem, value interface{}, variable string) []lsp.CompletionItem {
	var (
		itemKind      = lsp.CompletionItemKindVariable
		valueOf       = reflect.ValueOf(value)
		documentation = valueOf.String()
	)

	logger.Println("ValueKind: ", valueOf)

	switch valueOf.Kind() {
	case reflect.Slice, reflect.Map:
		itemKind = lsp.CompletionItemKindStruct
		documentation = h.toYAML(value)
	case reflect.Bool:
		itemKind = lsp.CompletionItemKindVariable
		documentation = h.getBoolType(value)
	case reflect.Float32, reflect.Float64:
		documentation = fmt.Sprintf("%.2f", valueOf.Float())
		itemKind = lsp.CompletionItemKindVariable
	case reflect.Invalid:
		documentation = "<Unknown>"
	default:
		itemKind = lsp.CompletionItemKindField
	}

	return append(items, lsp.CompletionItem{
		Label:         variable,
		InsertText:    variable,
		Documentation: documentation,
		Detail:        valueOf.Kind().String(),
		Kind:          itemKind,
	})
}

func (h *langHandler) toYAML(value interface{}) string {
	valBytes, _ := yaml.Marshal(value)
	return string(valBytes)
}

func (h *langHandler) getBoolType(value interface{}) string {
	if val, ok := value.(bool); ok && val {
		return "True"
	}
	return "False"
}

func getVariableCompletionItems(helmDocs []HelmDocumentation) (result []lsp.CompletionItem) {
	for _, item := range helmDocs {
		result = append(result, variableCompletionItem(item))
	}
	return result
}

func variableCompletionItem(helmDocumentation HelmDocumentation) lsp.CompletionItem {
	return lsp.CompletionItem{
		Label:         helmDocumentation.Name,
		InsertText:    helmDocumentation.Name,
		Detail:        helmDocumentation.Detail,
		Documentation: helmDocumentation.Doc,
		Kind:          lsp.CompletionItemKindVariable,
	}
}

func getFunctionCompletionItems(helmDocs []HelmDocumentation) (result []lsp.CompletionItem) {
	for _, item := range helmDocs {
		result = append(result, functionCompletionItem(item))
	}
	return result
}

func functionCompletionItem(helmDocumentation HelmDocumentation) lsp.CompletionItem {
	return lsp.CompletionItem{
		Label:         helmDocumentation.Name,
		InsertText:    helmDocumentation.Name,
		Detail:        helmDocumentation.Detail,
		Documentation: helmDocumentation.Doc,
		Kind:          lsp.CompletionItemKindFunction,
	}
}

func getTextCompletionItems(gotemplateSnippet []godocs.GoTemplateSnippet) (result []lsp.CompletionItem) {
	for _, item := range gotemplateSnippet {
		result = append(result, textCompletionItem(item))
	}
	return result
}

func textCompletionItem(gotemplateSnippet godocs.GoTemplateSnippet) lsp.CompletionItem {
	return lsp.CompletionItem{
		Label: gotemplateSnippet.Name,
		TextEdit: &lsp.TextEdit{
			Range:   lsp.Range{},
			NewText: gotemplateSnippet.Snippet,
		},
		Detail:           gotemplateSnippet.Detail,
		Documentation:    gotemplateSnippet.Doc,
		Kind:             lsp.CompletionItemKindText,
		InsertTextFormat: lsp.InsertTextFormatSnippet,
		FilterText:       gotemplateSnippet.Filter,
	}
}
