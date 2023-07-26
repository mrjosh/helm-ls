package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	gotemplate "github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) handleDefinition(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {

	logger.Println(fmt.Sprintf("Definition provider"))
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
		word               = doc.ValueAt(params.Position)
		splitted           = strings.Split(word, ".")
		variableSplitted   = []string{}
		position           lsp.Position
		definitionFilePath string
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

	logger.Println(fmt.Sprintf("Definition checking for word < %s >", word))

	switch variableSplitted[0] {
	case "Values":
		definitionFilePath = filepath.Join(h.rootURI.Filename(), "values.yaml")
		if len(variableSplitted) > 1 {
			position, err = h.getValueDefinition(variableSplitted[1:])
		}
	case "Chart":
		definitionFilePath = filepath.Join(h.rootURI.Filename(), "Chart.yaml")
		if len(variableSplitted) > 1 {
			position, err = h.getChartDefinition(variableSplitted[1:])
		}
	}

	if err == nil && definitionFilePath != "" {
		result := lsp.Location{
			URI:   "file://" + lsp.DocumentURI(definitionFilePath),
			Range: lsp.Range{Start: position},
		}

		return reply(ctx, result, err)
	}
	logger.Printf("Had no match for definition. Error: %v", err)
	return reply(ctx, nil, err)
}

// definitionAstParsing takes the current node
// depending on the node type it either returns the node that defines the current variable
// or the yaml selector for the current value
func (h *langHandler) definitionAstParsing(doc *lsplocal.Document, position lsp.Position) lsp.Location {
	var (
		currentNode   = lsplocal.NodeAtPosition(doc.Ast, position)
		pointToLoopUp = sitter.Point{
			Row:    position.Line,
			Column: position.Character,
		}
		relevantChildNode = lsplocal.FindRelevantChildNode(currentNode, pointToLoopUp)
	)

	switch relevantChildNode.Type() {
	case gotemplate.NodeTypeIdentifier:
		if relevantChildNode.Parent().Type() == gotemplate.NodeTypeVariable {

			variableName := relevantChildNode.Content([]byte(doc.Content))
			var node = lsplocal.GetVariableDefinition(variableName, relevantChildNode.Parent(), doc.Content)
			return lsp.Location{URI: doc.URI, Range: lsp.Range{Start: node.StartPoint(), End: node.EndPoint()}}
		}
	case gotemplate.NodeTypeDot, gotemplate.NodeTypeDotSymbol:
		return h.getDefinitionForValue(relevantChildNode, doc)
	}

	return lsp.Location{}
}

func (h *langHandler) getDefinitionForValue(node *sitter.Node, doc *lsplocal.Document) lsp.Location {
	var (
		yamlPathString     = getYamlPath(node, doc)
		yamlPath, err      = util.NewYamlPath(yamlPathString)
		definitionFilePath string
		position           lsp.Position
	)
	if err != nil {
		return lsp.Location{}
	}

	if yamlPath.IsValuesPath() {
		definitionFilePath = filepath.Join(h.rootURI.Filename(), "values.yaml")
		position, err = h.getValueDefinition(yamlPath.GetTail())
	}
	if yamlPath.IsChartPath() {
		definitionFilePath = filepath.Join(h.rootURI.Filename(), "Chart.yaml")
		position, err = h.getChartDefinition(yamlPath.GetTail())
	}

	if err == nil && definitionFilePath != "" {
		return lsp.Location{
			URI:   "file://" + lsp.DocumentURI(definitionFilePath),
			Range: lsp.Range{Start: position},
		}
	}
	return lsp.Location{}
}

func getYamlPath(node *sitter.Node, doc *lsplocal.Document) string {
	switch node.Type() {
	case gotemplate.NodeTypeDot:
		return lsplocal.TraverseIdentifierPathUp(node, doc)
	case gotemplate.NodeTypeDotSymbol:
		return lsplocal.GetFieldIdentifierPath(node, doc)
	default:
		return ""
	}
}

func (h *langHandler) getValueDefinition(splittedVar []string) (lsp.Position, error) {
	return util.GetPositionOfNode(h.valueNode, splittedVar)
}
func (h *langHandler) getChartDefinition(splittedVar []string) (lsp.Position, error) {

	modifyedVar := make([]string, 0)

	// for Releases, we make the first letter lowercase TODO: only do this when really needed
	for _, value := range splittedVar {
		restOfString := ""
		if (len(value)) > 1 {
			restOfString = value[1:]
		}
		firstLetterLowercase := strings.ToLower(string(value[0])) + restOfString
		modifyedVar = append(modifyedVar, firstLetterLowercase)
	}

	return util.GetPositionOfNode(h.chartNode, modifyedVar)
}
