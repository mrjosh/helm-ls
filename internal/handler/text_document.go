package handler

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/mrjosh/helm-ls/internal/charts"
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

	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	chart, err := h.chartStore.GetChartOrParentForDoc(doc.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", doc.URI, err)
	}
	notifications := lsplocal.GetDiagnosticsNotifications(chart, doc)

	defer h.publishDiagnostics(ctx, notifications)

	return nil
}

func (h *langHandler) DidClose(_ context.Context, _ *lsp.DidCloseTextDocumentParams) (err error) {
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
	notifications := lsplocal.GetDiagnosticsNotifications(chart, doc)

	defer h.publishDiagnostics(ctx, notifications)

	return nil
}

func (h *langHandler) DidChange(_ context.Context, params *lsp.DidChangeTextDocumentParams) (err error) {
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

func (h *langHandler) LoadDocsOnNewChart(chart *charts.Chart) {
	if chart.HelmChart == nil {
		return
	}

	for _, file := range chart.HelmChart.Templates {
		h.documents.Store(filepath.Join(chart.RootURI.Filename(), file.Name), file.Data, h.helmlsConfig)
	}

	for _, file := range chart.GetDependeciesTemplates() {
		logger.Debug(fmt.Sprintf("Storing dependency %s", file.Path))
		h.documents.Store(file.Path, file.Content, h.helmlsConfig)
	}
}
