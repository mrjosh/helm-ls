package templatehandler

import (
	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/log"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
)

var logger = log.GetLogger()

type TemplateHandler struct {
	documents       *lsplocal.DocumentStore
	chartStore      *charts.ChartStore
	yamllsConnector *yamlls.Connector
}

func NewTemplateHandler(documents *lsplocal.DocumentStore, chartStore *charts.ChartStore, yamllsConnector *yamlls.Connector) *TemplateHandler {
	return &TemplateHandler{
		documents:       documents,
		chartStore:      chartStore,
		yamllsConnector: yamllsConnector,
	}
}

func (h *TemplateHandler) SetChartStore(chartStore *charts.ChartStore) {
	h.chartStore = chartStore
}

func (h *TemplateHandler) SetYamllsConnector(yamllsConnector *yamlls.Connector) {
	h.yamllsConnector = yamllsConnector
}
