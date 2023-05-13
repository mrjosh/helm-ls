package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"strings"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	gotemplate "github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/pkg/chartutil"
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
	yaml "gopkg.in/yaml.v2"
)

var (
	emptyItems               = make([]lsp.CompletionItem, 0)
	functionsCompletionItems = make([]lsp.CompletionItem, 0)
)

func init() {
	functionsCompletionItems = append(functionsCompletionItems, getFunctionCompletionItems(helmFuncs)...)
	functionsCompletionItems = append(functionsCompletionItems, getFunctionCompletionItems(builtinFuncs)...)
	functionsCompletionItems = append(functionsCompletionItems, getFunctionCompletionItems(sprigFuncs)...)
}

func (h *langHandler) handleTextDocumentCompletion(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {

	if req.Params() == nil {
		return &jsonrpc2.Error{Code: jsonrpc2.InvalidParams}
	}

	var params lsp.CompletionParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}

	// logger.Println(params.Position.Character)
	// params.Position.Character = params.Position.Character - 1
	// logger.Println(params.Position.Character)
	word, err := completionAstParsing(doc, params.Position)

	if err != nil {

		logger.Println("Calling yamlls for completions")
		var response = reflect.New(reflect.TypeOf(lsp.CompletionList{})).Interface()
		_, err = h.yamllsConnector.Conn.Call(ctx, lsp.MethodTextDocumentCompletion, params, response)
		if err != nil {
			logger.Println("Error Calling yamlls for completions", err)
		}

		logger.Println("Got completions from yamlls", response)
		return reply(ctx, response, err)
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

	logger.Println(fmt.Sprintf("Word < %s >", word))

	if len(variableSplitted) == 0 {
		return reply(ctx, basicItems, err)
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
		items = h.getValue(h.values, variableSplitted[1:])
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

	return reply(ctx, items, err)
}

func completionAstParsing(doc *lsplocal.Document, position lsp.Position) (string, error) {
	var (
		currentNode   = lsplocal.NodeAtPosition(doc.Ast, position)
		pointToLoopUp = sitter.Point{
			Row:    position.Line,
			Column: position.Character,
		}
		relevantChildNode = lsplocal.FindRelevantChildNode(currentNode, pointToLoopUp)
		word              string
	)

	logger.Println("currentNode", currentNode)
	logger.Println("relevantChildNode", relevantChildNode.Type())

	switch relevantChildNode.Type() {
	case gotemplate.NodeTypeIdentifier:
		word = relevantChildNode.Content([]byte(doc.Content))
	case gotemplate.NodeTypeDot:
		logger.Println("TraverseIdentifierPathUp")
		word = lsplocal.TraverseIdentifierPathUp(relevantChildNode, doc)
	case gotemplate.NodeTypeDotSymbol:
		logger.Println("GetFieldIdentifierPath")
		word = lsplocal.GetFieldIdentifierPath(relevantChildNode, doc)
	}
	return word, nil
}

func findRelevantChildNode(currentNode *sitter.Node, pointToLookUp sitter.Point) *sitter.Node {
	for i := 0; i < int(currentNode.ChildCount()); i++ {
		child := currentNode.Child(i)
		if isPointLargerOrEq(pointToLookUp, child.StartPoint()) && isPointLargerOrEq(child.EndPoint(), pointToLookUp) {
			logger.Println("loop", child)
			return findRelevantChildNode(child, pointToLookUp)
		}
	}
	return currentNode
}

func isPointLarger(a sitter.Point, b sitter.Point) bool {
	if a.Row == b.Row {
		return a.Column > b.Column
	}
	return a.Row > b.Row
}
func isPointLargerOrEq(a sitter.Point, b sitter.Point) bool {
	if a.Row == b.Row {
		return a.Column >= b.Column
	}
	return a.Row > b.Row
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
