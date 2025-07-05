package yamlhandler

import (
	"context"
	"fmt"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/lsp/symboltable"
	"go.lsp.dev/protocol"
)

// Definition implements handler.LangHandler.
func (h *YamlHandler) Definition(ctx context.Context, params *protocol.DefinitionParams) (result []protocol.Location, err error) {
	logger.Debug("YamlHandler.Definition, for document at position", params.TextDocument.URI, params.Position)
	path, err := h.getYamlPath(params.TextDocument.URI, params.Position)
	if err != nil {
		return nil, fmt.Errorf("Getting References failed for document: %w", err)
	}
	templateContext := symboltable.TemplateContextFromYAMLPath(path)
	locations, err := h.getDefinitionsInValues(params.TextDocument.URI, templateContext)
	result = []protocol.Location{}

	// remove the current document
	for _, location := range locations {
		if location.URI != params.TextDocument.URI {
			result = append(result, location)
		}
	}

	return result, nil
}

func (h *YamlHandler) getDefinitionsInValues(uri protocol.URI, templateContext symboltable.TemplateContext) ([]protocol.Location, error) {
	chart, err := h.chartStore.GetChartForDoc(uri)
	if err != nil {
		return nil, fmt.Errorf("Getting Definitions failed for document, could not get chart for document: %w", err)
	}

	locations := []protocol.Location{}

	for _, value := range chart.ResolveValueFiles(templateContext, h.chartStore) {
		locs := value.ValuesFiles.GetPositionsForValue(value.Selector)
		if len(locs) > 0 {
			for _, valuesFile := range value.ValuesFiles.AllValuesFiles() {
				charts.SyncToDisk(valuesFile)
			}
		}

		locations = append(locations, locs...)

	}

	return locations, err
}
