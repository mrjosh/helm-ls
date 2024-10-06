package handler

import (
	"context"
	"errors"

	"github.com/mrjosh/helm-ls/internal/charts"
	helmlint "github.com/mrjosh/helm-ls/internal/helm_lint"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	lsp "go.lsp.dev/protocol"
)

func (h *ServerHandler) DidOpen(ctx context.Context, params *lsp.DidOpenTextDocumentParams) (err error) {
	if lsplocal.IsTemplateDocumentLangID(params.TextDocument.LanguageID) {
		doc, err := h.documents.DidOpenTemplateDocument(params, h.helmlsConfig)
		if err != nil {
			logger.Error(err)
			return err
		}

		h.yamllsConnector.DocumentDidOpen(doc.Ast, *params)

		chart, err := h.chartStore.GetChartOrParentForDoc(doc.URI)
		if err != nil {
			logger.Error("Error getting chart info for file", doc.URI, err)
		}

		defer h.publishDiagnostics(ctx, helmlint.GetDiagnosticsNotifications(chart, doc))

	}

	return nil
}

func (h *ServerHandler) DidClose(_ context.Context, _ *lsp.DidCloseTextDocumentParams) (err error) {
	return nil
}

func (h *ServerHandler) DidSave(ctx context.Context, params *lsp.DidSaveTextDocumentParams) (err error) {
	doc, ok := h.documents.GetTemplateDoc(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	chart, err := h.chartStore.GetChartOrParentForDoc(doc.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", doc.URI, err)
	}

	h.yamllsConnector.DocumentDidSave(doc, *params)
	notifications := helmlint.GetDiagnosticsNotifications(chart, doc)

	defer h.publishDiagnostics(ctx, notifications)

	return nil
}

func (h *ServerHandler) DidChange(_ context.Context, params *lsp.DidChangeTextDocumentParams) (err error) {
	doc, ok := h.documents.GetTemplateDoc(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}

	// Synchronise changes into the doc's ContentChanges
	doc.ApplyChanges(params.ContentChanges)

	shouldSendFullUpdateToYamlls := false
	for _, change := range params.ContentChanges {
		node := lsplocal.NodeAtPosition(doc.Ast, change.Range.Start)
		if node.Type() != "text" {
			shouldSendFullUpdateToYamlls = true
			break
		}
	}
	if shouldSendFullUpdateToYamlls {
		h.yamllsConnector.DocumentDidChangeFullSync(doc, *params)
	} else {
		h.yamllsConnector.DocumentDidChange(doc, *params)
	}

	return nil
}

func (h *ServerHandler) DidCreateFiles(ctx context.Context, params *lsp.CreateFilesParams) (err error) {
	logger.Error("DidCreateFiles unimplemented")
	return nil
}

// DidDeleteFiles implements protocol.Server.
func (h *ServerHandler) DidDeleteFiles(ctx context.Context, params *lsp.DeleteFilesParams) (err error) {
	logger.Error("DidDeleteFiles unimplemented")
	return nil
}

// DidRenameFiles implements protocol.Server.
func (h *ServerHandler) DidRenameFiles(ctx context.Context, params *lsp.RenameFilesParams) (err error) {
	logger.Error("DidRenameFiles unimplemented")
	return nil
}

func (h *ServerHandler) LoadDocsOnNewChart(chart *charts.Chart) {
	h.documents.LoadDocsOnNewChart(chart, h.helmlsConfig)
}
