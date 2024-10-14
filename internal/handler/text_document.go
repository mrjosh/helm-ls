package handler

import (
	"context"
	"errors"

	"github.com/mrjosh/helm-ls/internal/charts"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	lsp "go.lsp.dev/protocol"
)

func (h *ServerHandler) DidOpen(ctx context.Context, params *lsp.DidOpenTextDocumentParams) (err error) {
	handler := h.langHandlers[lsplocal.TemplateDocumentTypeForLangID(params.TextDocument.LanguageID)]

	if handler == nil {
		message := "Language not supported: " + string(params.TextDocument.LanguageID)
		logger.Error(message)
		return errors.New(message)
	}

	handler.DidOpen(ctx, params, h.helmlsConfig)

	defer h.publishDiagnostics(ctx, handler.GetDiagnostics(params.TextDocument.URI))

	return nil
}

func (h *ServerHandler) DidClose(_ context.Context, _ *lsp.DidCloseTextDocumentParams) (err error) {
	return nil
}

func (h *ServerHandler) DidSave(ctx context.Context, params *lsp.DidSaveTextDocumentParams) (err error) {
	handler, err := h.selectLangHandler(ctx, params.TextDocument.URI)
	if err != nil {
		return err
	}

	handler.DidSave(ctx, params)

	defer h.publishDiagnostics(ctx, handler.GetDiagnostics(params.TextDocument.URI))

	return nil
}

func (h *ServerHandler) DidChange(ctx context.Context, params *lsp.DidChangeTextDocumentParams) (err error) {
	handler, err := h.selectLangHandler(ctx, params.TextDocument.URI)
	handler.DidChange(ctx, params)

	doc, ok := h.documents.GetSyncDocument(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	// Synchronise changes into the doc's ContentChanges
	doc.ApplyChanges(params.ContentChanges)

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
