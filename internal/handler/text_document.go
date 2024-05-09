package handler

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (h *langHandler) DidOpen(ctx context.Context, params *lsp.DidOpenTextDocumentParams) (err error) {
	doc, err := h.documents.DidOpen(params, h.helmlsConfig)
	if err != nil {
		logger.Error(err)
		return err
	}

	h.yamllsConnector.DocumentDidOpen(doc.Ast, *params)

	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	chart, err := h.chartStore.GetChartOrParentForDoc(doc.URI)
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
	chart, err := h.chartStore.GetChartOrParentForDoc(doc.URI)
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

func (h *langHandler) DidCreateFiles(ctx context.Context, params *lsp.CreateFilesParams) (err error) {
	logger.Error("DidCreateFiles unimplemented")
	return nil
}

// DidDeleteFiles implements protocol.Server.
func (h *langHandler) DidDeleteFiles(ctx context.Context, params *lsp.DeleteFilesParams) (err error) {
	logger.Error("DidDeleteFiles unimplemented")
	return nil
}

// DidRenameFiles implements protocol.Server.
func (h *langHandler) DidRenameFiles(ctx context.Context, params *lsp.RenameFilesParams) (err error) {
	logger.Error("DidRenameFiles unimplemented")
	return nil
}

// TODO: maybe use the helm implementation of this once https://github.com/mrjosh/helm-ls/pull/77 is resolved
func (h *langHandler) LoadDocsOnNewChart(rootURI uri.URI) {
	_ = filepath.WalkDir(filepath.Join(rootURI.Filename(), "templates"),
		func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() {
				return h.documents.Store(uri.File(path), h.helmlsConfig)
			}
			return nil
		},
	)
}
