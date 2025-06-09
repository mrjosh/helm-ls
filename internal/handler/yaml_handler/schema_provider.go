package yamlhandler

import (
	"context"
	"fmt"

	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"go.lsp.dev/uri"
)

func (h *YamlHandler) CustomSchemaProvider(ctx context.Context, URI uri.URI) (uri.URI, error) {
	if h.jsonSchemas == nil {
		return uri.New(""), fmt.Errorf("JSON schema generator not initialized")
	}

	if !document.IsValuesYamlFile(URI) {
		return uri.New(""), nil
	}

	chart, err := h.chartStore.GetChartForDoc(URI)
	if err != nil {
		logger.Error("Could not get a chart for the document: ", err)
		// we can ignore the error, providing a wrong schema is still useful
		// chart will still include some fallback values
	}

	schemaFilePath, err := h.jsonSchemas.GetJSONSchemaForChart(chart)
	if err != nil {
		logger.Error(err)
		return uri.New(""), err
	}
	return uri.File(schemaFilePath), nil
}
