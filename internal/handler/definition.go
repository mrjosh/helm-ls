package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/mrjosh/helm-ls/internal/charts"
	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	gotemplate "github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
	"gopkg.in/yaml.v3"
)

func (h *langHandler) Definition(ctx context.Context, params *lsp.DefinitionParams) (result []lsp.Location, err error) {
	genericDocumentUseCase, err := h.NewGenericDocumentUseCase(params.TextDocumentPositionParams)
	if err != nil {
		return nil, err
	}
	doc := genericDocumentUseCase.Document
	chart := genericDocumentUseCase.Chart

	result, err = h.definitionAstParsing(genericDocumentUseCase, chart, doc, params.Position)
	if err != nil {
		// suppress errors for clients
		// otherwise using go-to-definition on words that have no definition
		// will result in an error
		logger.Println("Error getting definitions", err)
		return nil, nil
	}
	return result, nil
}

func (h *langHandler) definitionAstParsing(genericDocumentUseCase *languagefeatures.GenericDocumentUseCase, chart *charts.Chart, doc *lsplocal.Document, position lsp.Position) ([]lsp.Location, error) {
	var (
		relevantChildNode = genericDocumentUseCase.Node
		parentNode        = relevantChildNode.Parent()
		parentType        string
	)

	if parentNode != nil {
		parentType = parentNode.Type()
	}

	nodeType := relevantChildNode.Type()
	switch nodeType {
	case gotemplate.NodeTypeIdentifier:
		logger.Println("Parent type", relevantChildNode.Parent().Type())
		if parentType == gotemplate.NodeTypeVariable {
			return h.getDefinitionForVariable(relevantChildNode, doc)
		}

		if parentType == gotemplate.NodeTypeSelectorExpression || parentType == gotemplate.NodeTypeField {
			return h.getDefinitionForValue(chart, relevantChildNode, doc)
		}
		return h.getDefinitionForFixedIdentifier(chart, relevantChildNode, doc)
	case gotemplate.NodeTypeDot, gotemplate.NodeTypeDotSymbol, gotemplate.NodeTypeFieldIdentifier:
		return h.getDefinitionForValue(chart, relevantChildNode, doc)
	}

	if parentType == gotemplate.NodeTypeArgumentList {
		includesCallFeature := languagefeatures.NewIncludesCallFeature(genericDocumentUseCase)
		return includesCallFeature.Definition()
	}

	return []lsp.Location{}, fmt.Errorf("Definition not implemented for node type %s", relevantChildNode.Type())
}

func (h *langHandler) getDefinitionForVariable(node *sitter.Node, doc *lsplocal.Document) ([]lsp.Location, error) {
	variableName := node.Content([]byte(doc.Content))
	definitionNode := lsplocal.GetVariableDefinition(variableName, node.Parent(), doc.Content)
	if definitionNode == nil {
		return []lsp.Location{}, fmt.Errorf("Could not find definition for %s. Variable definition not found", variableName)
	}
	return []lsp.Location{{URI: doc.URI, Range: lsp.Range{Start: util.PointToPosition(definitionNode.StartPoint())}}}, nil
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

	return []lsp.Location{}, fmt.Errorf("Could not find definition for %s. Fixed identifier not found", name)
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

	if definitionFileURI != "" {
		locations := []lsp.Location{}
		for _, position := range positions {
			locations = append(locations, lsp.Location{
				URI:   definitionFileURI,
				Range: lsp.Range{Start: position},
			})
		}
		return locations, nil
	}
	return []lsp.Location{}, fmt.Errorf("Could not find definition for %s. No definition found", yamlPath)
}

func getYamlPath(node *sitter.Node, doc *lsplocal.Document) string {
	switch node.Type() {
	case gotemplate.NodeTypeDot:
		return lsplocal.TraverseIdentifierPathUp(node, doc)
	case gotemplate.NodeTypeDotSymbol, gotemplate.NodeTypeFieldIdentifier, gotemplate.NodeTypeIdentifier:
		return lsplocal.GetFieldIdentifierPath(node, doc)
	default:
		logger.Error("Could not get yaml path for node type ", node.Type())
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
