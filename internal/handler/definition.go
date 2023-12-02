package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	gotemplate "github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) handleDefinition(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {
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
	result, err := h.definitionAstParsing(doc, params.Position)

	if err != nil {
		// suppress errors for clients
		// otherwise using go-to-definition on words that have no definition
		// will result in an error
		logger.Println(err)
		return reply(ctx, nil, nil)
	}
	return reply(ctx, result, err)
}

func (h *langHandler) definitionAstParsing(doc *lsplocal.Document, position lsp.Position) (lsp.Location, error) {
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
			return h.getDefinitionForVariable(relevantChildNode, doc)
		}
		return h.getDefinitionForFixedIdentifier(relevantChildNode, doc)
	case gotemplate.NodeTypeDot, gotemplate.NodeTypeDotSymbol, gotemplate.NodeTypeFieldIdentifier:
		return h.getDefinitionForValue(relevantChildNode, doc)
	}

	return lsp.Location{}, fmt.Errorf("Definition not implemented for node type %s", relevantChildNode.Type())
}

func (h *langHandler) getDefinitionForVariable(node *sitter.Node, doc *lsplocal.Document) (lsp.Location, error) {
	variableName := node.Content([]byte(doc.Content))
	var defintionNode = lsplocal.GetVariableDefinition(variableName, node.Parent(), doc.Content)
	if defintionNode == nil {
		return lsp.Location{}, fmt.Errorf("Could not find definition for %s", variableName)
	}
	return lsp.Location{URI: doc.URI, Range: lsp.Range{Start: util.PointToPosition(defintionNode.StartPoint())}}, nil
}

// getDefinitionForFixedIdentifier checks if the current identifier has a constant definition and returns it
func (h *langHandler) getDefinitionForFixedIdentifier(node *sitter.Node, doc *lsplocal.Document) (lsp.Location, error) {
	var name = node.Content([]byte(doc.Content))
	switch name {
	case "Values":
		return lsp.Location{
			URI: h.projectFiles.GetValuesFileURI()}, nil
	case "Chart":
		return lsp.Location{
			URI: h.projectFiles.GetChartFileURI()}, nil
	}

	return lsp.Location{}, fmt.Errorf("Could not find definition for %s", name)
}

func (h *langHandler) getDefinitionForValue(node *sitter.Node, doc *lsplocal.Document) (lsp.Location, error) {
	var (
		yamlPathString    = getYamlPath(node, doc)
		yamlPath, err     = util.NewYamlPath(yamlPathString)
		definitionFileURI lsp.DocumentURI
		position          lsp.Position
	)
	if err != nil {
		return lsp.Location{}, err
	}

	if yamlPath.IsValuesPath() {
		definitionFileURI = h.projectFiles.GetValuesFileURI()
		position, err = h.getValueDefinition(yamlPath.GetTail())
	}
	if yamlPath.IsChartPath() {
		definitionFileURI = h.projectFiles.GetChartFileURI()
		position, err = h.getChartDefinition(yamlPath.GetTail())
	}

	if err == nil && definitionFileURI != "" {
		return lsp.Location{
			URI:   definitionFileURI,
			Range: lsp.Range{Start: position},
		}, nil
	}
	return lsp.Location{}, fmt.Errorf("Could not find definition for %s", yamlPath)
}

func getYamlPath(node *sitter.Node, doc *lsplocal.Document) string {
	switch node.Type() {
	case gotemplate.NodeTypeDot:
		return lsplocal.TraverseIdentifierPathUp(node, doc)
	case gotemplate.NodeTypeDotSymbol, gotemplate.NodeTypeFieldIdentifier:
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
	// for Charts, we make the first letter lowercase
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
