package helmlint

import (
	"fmt"

	"github.com/goccy/go-yaml/ast"
	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"github.com/mrjosh/helm-ls/internal/lsp/symboltable"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

func LintUnusedValues(chart *charts.Chart, doc *document.YamlDocument, templateDocs []*document.TemplateDocument) []lsp.Diagnostic {
	if len(templateDocs) == 0 {
		// TODO: delay linting until the template docs are loaded
		return []lsp.Diagnostic{}
	}

	result := []lsp.Diagnostic{}
	for _, node := range FindNodes(doc.GoccyYamlNode) {

		templateContext := symboltable.TemplateContextFromYAMLPath(node.GetPath())
		found := false

		for _, templateDoc := range templateDocs {
			// TODO(dependecy-charts): template context would need to be adjusted for dependency charts
			// see https://github.com/mrjosh/helm-ls/issues/152
			referenceRanges := templateDoc.SymbolTable.GetTemplateContextRanges(append([]string{"Values"}, templateContext...))

			logger.Println(fmt.Sprintf("LintUnusedValues: checking template context %v in template doc %s", templateContext, templateDoc.GetPath()))

			if len(referenceRanges) > 0 {
				found = true
				break
			}
		}

		if found {
			continue
		}

		fmt.Println(node.String())
		result = append(result, lsp.Diagnostic{
			Range:           util.TokenToRange(node.GetToken()),
			Severity:        0,
			Code:            nil,
			CodeDescription: &lsp.CodeDescription{},
			Source:          "",
			Message:         fmt.Sprintf("Unused value: %s of type %s", node.GetPath(), node.Type()),
			Tags: []lsp.DiagnosticTag{
				lsp.DiagnosticTagUnnecessary,
			},
			RelatedInformation: []lsp.DiagnosticRelatedInformation{},
			Data:               nil,
		})
	}

	return result
}

func FindNodes(node ast.Node) []ast.Node {
	visitor := &LeafMappingsFinderVisitor{}
	ast.Walk(visitor, node)
	return visitor.result
}

// PositionFinderVisitor is a visitor that collects positions.
type LeafMappingsFinderVisitor struct {
	result []ast.Node
}

func (v *LeafMappingsFinderVisitor) Visit(node ast.Node) ast.Visitor {
	if IsLeafMapping(node) {
		v.result = append(v.result, node)
	}
	return v
}

func IsLeafMapping(node ast.Node) bool {
	switch node.Type() {
	case ast.MappingKeyType, ast.MappingValueType, ast.MappingType, ast.SequenceType:
		return false
	default:
		return true
	}
}
