package yamlhandler

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/pkg/errors"
	"go.lsp.dev/protocol"
)

// References implements handler.LangHandler.
func (h *YamlHandler) References(ctx context.Context, params *protocol.ReferenceParams) (result []protocol.Location, err error) {
	path, err := h.GetYamlPath(params.TextDocument.URI, params.Position)
	templateContext := util.JSONPathToTemplateContext(path)

	logger.Debug("YamlHandler References looking for template context", templateContext)
	locations := []protocol.Location{}
	for _, doc := range h.documents.GetAllTemplateDocs() {
		referenceRanges := doc.SymbolTable.GetTemplateContextRanges(append([]string{"Values"}, templateContext...))

		locations = append(locations, util.RangesToLocations(doc.URI, referenceRanges)...)
	}

	chart, err := h.chartStore.GetChartForDoc(params.TextDocument.URI)
	if err != nil {
		return locations, errors.Wrap(err, "could not get chart for document")
	}

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
