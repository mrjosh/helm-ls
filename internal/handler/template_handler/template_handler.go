package templatehandler

import (
	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/log"
	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/protocol"
)

var logger = log.GetLogger()

type TemplateHandler struct {
	client          protocol.Client
	documents       *document.DocumentStore
	chartStore      *charts.ChartStore
	yamllsConnector *yamlls.Connector
	helmlintConfig  util.HelmLintConfig
}

func NewTemplateHandler(client protocol.Client, documents *document.DocumentStore, chartStore *charts.ChartStore, helmlintConfig util.HelmLintConfig) *TemplateHandler {
	return &TemplateHandler{
		client:          client,
		documents:       documents,
		chartStore:      chartStore,
		yamllsConnector: &yamlls.Connector{},
		helmlintConfig:  helmlintConfig,
	}
}

func (h *TemplateHandler) SetClient(client protocol.Client) {
	h.client = client
}

func (h *TemplateHandler) setYamllsConnector(yamllsConnector *yamlls.Connector) {
	h.yamllsConnector = yamllsConnector
}
