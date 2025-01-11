package yamlhandler

import (
	"context"

	jsonschema "github.com/mrjosh/helm-ls/internal/json_schema"
	"go.lsp.dev/uri"
)

func (h *YamlHandler) CustomSchemaProvider(ctx context.Context, URI uri.URI) (uri.URI, error) {
	chart, err := h.chartStore.GetChartForDoc(URI)
	if err != nil {
		logger.Error(err)
		// we can ignore the error, providing a wrong schema is still useful
		// chart will still include some fallback values
	}

	// TODO: we want to cache jsonschema generation, either here or in the jsonschema package

	schemaFilePath, err := jsonschema.CreateJsonSchemaForChart(chart)
	if err != nil {
		logger.Error(err)
		return uri.New(""), err
	}
	return uri.File(schemaFilePath), nil
}
