package handler

import (
	"context"
	"errors"

	lspinternal "github.com/mrjosh/helm-ls/internal/lsp"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) DidOpen(ctx context.Context, params *lsp.DidOpenTextDocumentParams) (err error) {
	doc, err := h.documents.DidOpen(params, h.helmlsConfig)
	if err != nil {
		logger.Error(err)
		return err
	}

	h.yamllsConnector.DocumentDidOpen(doc.Ast, *params)

	_, err = h.chartStore.GetChartForDoc(doc.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", doc.URI, err)
	}

	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	chart, err := h.chartStore.GetChartForDoc(doc.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", doc.URI, err)
	}
	notification := lsplocal.GetDiagnosticsNotification(chart, doc)

	return h.client.PublishDiagnostics(ctx, notification)
}

func (h *langHandler) DidClose(ctx context.Context, params *lsp.DidCloseTextDocumentParams) (err error) {
	return nil
}

func (h *langHandler) DidSave(ctx context.Context, params *lsp.DidSaveTextDocumentParams) (err error) {
	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	chart, err := h.chartStore.GetChartForDoc(doc.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", doc.URI, err)
	}

	h.yamllsConnector.DocumentDidSave(doc, *params)
	notification := lsplocal.GetDiagnosticsNotification(chart, doc)

	return h.client.PublishDiagnostics(ctx, notification)
}

func (h *langHandler) DidChange(ctx context.Context, params *lsp.DidChangeTextDocumentParams) (err error) {
	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}

	shouldSendFullUpdateToYamlls := false

	// Synchronise changes into the doc's ContentChanges
	doc.ApplyChanges(params.ContentChanges)

	for _, change := range params.ContentChanges {
		node := lspinternal.NodeAtPosition(doc.Ast, change.Range.Start)
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

func (h *langHandler) DidCreateFiles(ctx context.Context, params *lsp.CreateFilesParams) (err error) {
	panic("unimplemented")
}

// DidDeleteFiles implements protocol.Server.
func (h *langHandler) DidDeleteFiles(ctx context.Context, params *lsp.DeleteFilesParams) (err error) {
	panic("unimplemented")
}

// DidRenameFiles implements protocol.Server.
func (h *langHandler) DidRenameFiles(ctx context.Context, params *lsp.RenameFilesParams) (err error) {
	panic("unimplemented")
}