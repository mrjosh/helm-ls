package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mrjosh/helm-ls/internal/charts"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	gotemplate "github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
	"gopkg.in/yaml.v3"
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
	chart, err := h.chartStore.GetChartForDoc(params.TextDocument.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", params.TextDocument.URI, err)
	}

	result, err := h.definitionAstParsing(chart, doc, params.Position)
	if err != nil {
		// suppress errors for clients
		// otherwise using go-to-definition on words that have no definition
		// will result in an error
		logger.Println("Error getting definitions", err)
		return reply(ctx, nil, nil)
	}
	return reply(ctx, result, err)
}

func (h *langHandler) definitionAstParsing(chart *charts.Chart, doc *lsplocal.Document, position lsp.Position) ([]lsp.Location, error) {
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
		return h.getDefinitionForFixedIdentifier(chart, relevantChildNode, doc)
	case gotemplate.NodeTypeDot, gotemplate.NodeTypeDotSymbol, gotemplate.NodeTypeFieldIdentifier:
		return h.getDefinitionForValue(chart, relevantChildNode, doc)
	}

	return []lsp.Location{}, fmt.Errorf("Definition not implemented for node type %s", relevantChildNode.Type())
}

func (h *langHandler) getDefinitionForVariable(node *sitter.Node, doc *lsplocal.Document) ([]lsp.Location, error) {
	variableName := node.Content([]byte(doc.Content))
	defintionNode := lsplocal.GetVariableDefinition(variableName, node.Parent(), doc.Content)
	if defintionNode == nil {
		return []lsp.Location{}, fmt.Errorf("Could not find definition for %s", variableName)
	}
	return []lsp.Location{{URI: doc.URI, Range: lsp.Range{Start: util.PointToPosition(defintionNode.StartPoint())}}}, nil
}

// getDefinitionForFixedIdentifier checks if the current identifier has a constant definition and returns it
func (h *langHandler) getDefinitionForFixedIdentifier(chart *charts.Chart, node *sitter.Node, doc *lsplocal.Document) ([]lsp.Location, error) {
	name := node.Content([]byte(doc.Content))
	switch name {
	case "Values":
		result := []lsp.Location{}

		for _, valueFile := range chart.ValuesFiles.AllValuesFiles() {
			result = append(result, lsp.Location{URI: valueFile.URI})
		}
		return result, nil

	case "Chart":
		return []lsp.Location{
				{URI: chart.ChartMetadata.URI},
			},
			nil
	}

	return []lsp.Location{}, fmt.Errorf("Could not find definition for %s", name)
}

func (h *langHandler) getDefinitionForValue(chart *charts.Chart, node *sitter.Node, doc *lsplocal.Document) ([]lsp.Location, error) {
	var (
		yamlPathString    = getYamlPath(node, doc)
		yamlPath, err     = util.NewYamlPath(yamlPathString)
		definitionFileURI lsp.DocumentURI
		positions         []lsp.Position
	)
	if err != nil {
		return []lsp.Location{}, err
	}

	if yamlPath.IsValuesPath() {
		return h.getValueDefinition(chart, yamlPath.GetTail()), nil
	}
	if yamlPath.IsChartPath() {
		definitionFileURI = chart.ChartMetadata.URI
		position, err := h.getChartDefinition(&chart.ChartMetadata.YamlNode, yamlPath.GetTail())
		if err == nil {
			positions = append(positions, position)
		}
	}

	if err == nil && definitionFileURI != "" {
		locations := []lsp.Location{}
		for _, position := range positions {
			locations = append(locations, lsp.Location{
				URI:   definitionFileURI,
				Range: lsp.Range{Start: position},
			})
		}
		return locations, nil
	}
	return []lsp.Location{}, fmt.Errorf("Could not find definition for %s", yamlPath)
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

func (h *langHandler) getValueDefinition(chart *charts.Chart, splittedVar []string) []lsp.Location {
	locations := []lsp.Location{}
	for _, value := range chart.ResolveValueFiles(splittedVar, h.chartStore) {
		locations = append(locations, value.ValuesFiles.GetPositionsForValue(value.Selector)...)
	}
	return locations
}

func (h *langHandler) getChartDefinition(chartNode *yaml.Node, splittedVar []string) (lsp.Position, error) {
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
	return util.GetPositionOfNode(chartNode, modifyedVar)
}
