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

func definitionAstParsing(doc *lsplocal.Document, position lsp.Position) string {
	var (
		currentNode   = lsplocal.NodeAtPosition(doc.Ast, position)
		pointToLoopUp = sitter.Point{
			Row:    position.Line,
			Column: position.Character,
		}
		relevantChildNode = lsplocal.FindRelevantChildNode(currentNode, pointToLoopUp)
		word              string
	)

	switch relevantChildNode.Type() {
	case gotemplate.NodeTypeIdentifier:
		if relevantChildNode.Parent().Type() == gotemplate.NodeTypeVariable {

			variableName := relevantChildNode.Content([]byte(doc.Content))
			lsplocal.GetVariableDefinition(variableName, relevantChildNode.Parent(), doc)
		}
	case gotemplate.NodeTypeDot:
		word = lsplocal.TraverseIdentifierPathUp(relevantChildNode, doc)
	case gotemplate.NodeTypeDotSymbol:
		word = lsplocal.GetFieldIdentifierPath(relevantChildNode, doc)
	}

	return word

}

func (h *langHandler) getValueDefinition(splittedVar []string) (lsp.Position, error) {
	return util.GetPositionOfNode(h.valueNode, splittedVar)
}
func (h *langHandler) getChartDefinition(splittedVar []string) (lsp.Position, error) {

	modifyedVar := make([]string, 0)

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
