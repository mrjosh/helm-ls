package templatehandler

import (
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	lsp "go.lsp.dev/protocol"
)

func (h *TemplateHandler) GetDiagnostics(uri lsp.DocumentURI) []lsp.PublishDiagnosticsParams {
	doc, ok := h.documents.GetTemplateDoc(uri)
	if !ok {
		logger.Error("Could not get document: " + uri.Filename())
		return []lsp.PublishDiagnosticsParams{}
	}
	chart, err := h.chartStore.GetChartOrParentForDoc(doc.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", doc.URI, err)
	}
	notifications := lsplocal.GetDiagnosticsNotifications(chart, doc)
	return notifications
}
