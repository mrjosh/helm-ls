package yamlhandler

import (
	"context"
	"fmt"

	"github.com/mrjosh/helm-ls/internal/lsp/symboltable"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/protocol"
)

// References implements handler.LangHandler.
func (h *YamlHandler) References(ctx context.Context, params *protocol.ReferenceParams) (result []protocol.Location, err error) {
	path, err := h.getYamlPath(params.TextDocument.URI, params.Position)
	if err != nil {
		return nil, fmt.Errorf("Getting References failed for document, could not get YAML Path: %w", err)
	}
	templateContext := symboltable.TemplateContextFromYAMLPath(path)

	logger.Debug("YamlHandler References looking for template context", templateContext)

	locations := h.getReferencesInTemplates(templateContext)

	definitions, err := h.getDefinitionsInValues(params.TextDocument.URI, templateContext)
	if err != nil {
		return locations, fmt.Errorf("Adding definitions failed to references failed: %w", err)
	}
	locations = append(locations, definitions...)

	return locations, nil
}

func (h *YamlHandler) getReferencesInTemplates(templateContext symboltable.TemplateContext) []protocol.Location {
	locations := []protocol.Location{}
	for _, doc := range h.documents.GetAllTemplateDocs() {
		// TODO(dependecy-charts): template context would need to be adjusted for dependency charts
		// see https://github.com/mrjosh/helm-ls/issues/152
		referenceRanges := doc.SymbolTable.GetTemplateContextRanges(append([]string{"Values"}, templateContext...))

		locations = append(locations, util.RangesToLocations(doc.URI, referenceRanges)...)
	}
	return locations
}
