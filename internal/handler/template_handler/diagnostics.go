package templatehandler

import (
	helmlint "github.com/mrjosh/helm-ls/internal/helm_lint"
	lsp "go.lsp.dev/protocol"
)

func (h *TemplateHandler) GetDiagnostics(uri lsp.DocumentURI) []lsp.PublishDiagnosticsParams {
	logger.Debug("TemplateHandler.GetDiagnostics")
	doc, ok := h.documents.GetTemplateDoc(uri)
	if !ok {
		logger.Error("Could not get document: " + uri.Filename())
		return []lsp.PublishDiagnosticsParams{}
	}
	chart, err := h.chartStore.GetChartOrParentForDoc(doc.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", doc.URI, err)
	}
	if chart == nil {
		return []lsp.PublishDiagnosticsParams{}
	}
	notifications := helmlint.GetDiagnosticsNotifications(chart, doc)
	return notifications
}
