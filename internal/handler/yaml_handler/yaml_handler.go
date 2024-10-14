package yamlhandler

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/log"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

var logger = log.GetLogger()

type YamlHandler struct {
	documents       *lsplocal.DocumentStore
	chartStore      *charts.ChartStore
	yamllsConnector *yamlls.Connector
}

// Completion implements handler.LangHandler.
func (h *YamlHandler) Completion(ctx context.Context, params *protocol.CompletionParams) (result *protocol.CompletionList, err error) {
	panic("unimplemented")
}

// Configure implements handler.LangHandler.
func (h *YamlHandler) Configure(ctx context.Context, helmlsConfig util.HelmlsConfiguration) {}

// Definition implements handler.LangHandler.
func (h *YamlHandler) Definition(ctx context.Context, params *protocol.DefinitionParams) (result []protocol.Location, err error) {
	panic("unimplemented")
}

// DocumentSymbol implements handler.LangHandler.
func (h *YamlHandler) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) (result []interface{}, err error) {
	panic("unimplemented")
}

// GetDiagnostics implements handler.LangHandler.
func (h *YamlHandler) GetDiagnostics(uri uri.URI) []protocol.PublishDiagnosticsParams {
	panic("unimplemented")
}

// Hover implements handler.LangHandler.
func (h *YamlHandler) Hover(ctx context.Context, params *protocol.HoverParams) (result *protocol.Hover, err error) {
	panic("unimplemented")
}

// References implements handler.LangHandler.
func (h *YamlHandler) References(ctx context.Context, params *protocol.ReferenceParams) (result []protocol.Location, err error) {
	panic("unimplemented")
}

// SetClient implements handler.LangHandler.
func (h *YamlHandler) SetClient(client protocol.Client) {}

func NewYamlHandler(client protocol.Client, documents *lsplocal.DocumentStore, chartStore *charts.ChartStore) *YamlHandler {
	return &YamlHandler{
		documents:       documents,
		chartStore:      chartStore,
		yamllsConnector: &yamlls.Connector{},
	}
}

func (h *YamlHandler) SetChartStore(chartStore *charts.ChartStore) {
	h.chartStore = chartStore
}
