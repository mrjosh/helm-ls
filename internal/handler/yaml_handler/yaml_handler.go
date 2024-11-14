package yamlhandler

import (
	"context"
	"encoding/json"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	jsonschema "github.com/mrjosh/helm-ls/internal/json_schema"
	"github.com/mrjosh/helm-ls/internal/log"
	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

var logger = log.GetLogger()

type YamlHandler struct {
	documents       *document.DocumentStore
	chartStore      *charts.ChartStore
	client          protocol.Client
	yamllsConnector *yamlls.Connector
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

func NewYamlHandler(client protocol.Client, documents *document.DocumentStore, chartStore *charts.ChartStore) *YamlHandler {
	return &YamlHandler{
		documents:       documents,
		chartStore:      chartStore,
		yamllsConnector: &yamlls.Connector{},
	}
}

func (h *YamlHandler) SetChartStore(chartStore *charts.ChartStore) {
	h.chartStore = chartStore
}

func (h *YamlHandler) setYamllsConnector(yamllsConnector *yamlls.Connector) {
	h.yamllsConnector = yamllsConnector
}

func (h *YamlHandler) CustomSchemaProvider(ctx context.Context, URI uri.URI) (uri.URI, error) {
	chart, err := h.chartStore.GetChartForDoc(URI)
	if err != nil {
		logger.Error(err)
	}
	schemaFilePath, err := jsonschema.CreateJsonSchemaForChart(chart)
	if err != nil {
		logger.Error(err)
		return uri.New(""), err
	}
	return uri.File(schemaFilePath), nil
}

func (h *YamlHandler) CustomHandler(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	switch req.Method() {
	case "custom/schema/request":

		params := []string{}
		jsonBytes, err := req.Params().MarshalJSON()
		if err != nil {
			logger.Error(err)
			return reply(ctx, nil, nil)
		}

		err = json.Unmarshal(jsonBytes, &params)
		if err != nil {
			logger.Error(err)
			return reply(ctx, nil, nil)
		}

		logger.Println("YamlHandler: custom/schema/request", req.Params())

		if len(params) == 0 {
			return reply(ctx, nil, nil)
		}
		chart, err := h.chartStore.GetChartForDoc(uri.New(params[0]))
		if err != nil {
			logger.Error(err)
		}
		schemaFilePath, err := jsonschema.CreateJsonSchemaForChart(chart)
		if err != nil {
			logger.Error(err)
			return reply(ctx, nil, nil)
		}

		return reply(ctx, uri.New(schemaFilePath), nil)
	}

	return jsonrpc2.MethodNotFoundHandler(ctx, reply, req)
}
