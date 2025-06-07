package handler

import (
	"context"
	"errors"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/lsp/document"
	lsp "go.lsp.dev/protocol"
)

func (h *ServerHandler) DidOpen(ctx context.Context, params *lsp.DidOpenTextDocumentParams) (err error) {
	handler := h.langHandlers[document.TemplateDocumentTypeForLangID(params.TextDocument.LanguageID)]

	if handler == nil {
		message := "Language not supported: " + string(params.TextDocument.LanguageID)
		logger.Error(message)
		return errors.New(message)
	}

	logger.Debug("Running DidOpen for language ", params.TextDocument.LanguageID)
	defer func() {
		h.publishDiagnostics(ctx, handler.GetDiagnostics(params.TextDocument.URI))
	}()
	return handler.DidOpen(ctx, params, h.helmlsConfig)
}

func (h *ServerHandler) DidClose(_ context.Context, _ *lsp.DidCloseTextDocumentParams) (err error) {
	return nil
}

func (h *ServerHandler) DidSave(ctx context.Context, params *lsp.DidSaveTextDocumentParams) (err error) {
	handler, err := h.selectLangHandler(ctx, params.TextDocument.URI)
	if err != nil {
		return err
	}

	defer h.publishDiagnostics(ctx, handler.GetDiagnostics(params.TextDocument.URI))
	return handler.DidSave(ctx, params)
}

func (h *ServerHandler) DidChange(ctx context.Context, params *lsp.DidChangeTextDocumentParams) (err error) {
	doc, ok := h.documents.GetSyncDocument(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	// Synchronise changes into the doc's ContentChanges
	doc.ApplyChanges(params.ContentChanges)

	handler, err := h.selectLangHandler(ctx, params.TextDocument.URI)
	if err != nil {
		return err
	}
	return handler.PostDidChange(ctx, params)
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
