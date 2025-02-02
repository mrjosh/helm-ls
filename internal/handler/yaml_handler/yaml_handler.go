package yamlhandler

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	jsonschema "github.com/mrjosh/helm-ls/internal/json_schema"
	"github.com/mrjosh/helm-ls/internal/log"
	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"go.lsp.dev/protocol"
)

var logger = log.GetLogger()

type YamlHandler struct {
	documents       *document.DocumentStore
	chartStore      *charts.ChartStore
	client          protocol.Client
	yamllsConnector *yamlls.Connector
	jsonSchemas     *jsonschema.JSONSchemaCache
}

// Definition implements handler.LangHandler.
func (h *YamlHandler) Definition(ctx context.Context, params *protocol.DefinitionParams) (result []protocol.Location, err error) {
	panic("unimplemented")
}

// References implements handler.LangHandler.
func (h *YamlHandler) References(ctx context.Context, params *protocol.ReferenceParams) (result []protocol.Location, err error) {
	panic("unimplemented")
}

// SetClient implements handler.LangHandler.
func (h *YamlHandler) SetClient(client protocol.Client) {}

func NewYamlHandler(client protocol.Client, documents *document.DocumentStore, chartStore *charts.ChartStore, jsonSchemas *jsonschema.JSONSchemaCache) *YamlHandler {
	return &YamlHandler{
		documents:       documents,
		chartStore:      chartStore,
		yamllsConnector: &yamlls.Connector{},
		jsonSchemas:     jsonSchemas,
	}
}

func (h *YamlHandler) SetChartStore(chartStore *charts.ChartStore) {
	h.chartStore = chartStore

	jsonSchemas := jsonschema.NewJSONSchemaCache(chartStore)
	h.jsonSchemas = jsonSchemas
}

func (h *YamlHandler) setYamllsConnector(yamllsConnector *yamlls.Connector) {
	h.yamllsConnector = yamllsConnector
}
