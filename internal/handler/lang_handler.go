package handler

import (
	"context"
	"fmt"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/protocol"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

type LangHandler interface {
	Completion(ctx context.Context, params *lsp.CompletionParams) (result *lsp.CompletionList, err error)
	References(ctx context.Context, params *lsp.ReferenceParams) (result []lsp.Location, err error)
	Hover(ctx context.Context, params *lsp.HoverParams) (result *lsp.Hover, err error)
	Definition(ctx context.Context, params *lsp.DefinitionParams) (result []lsp.Location, err error)
	DocumentSymbol(ctx context.Context, params *lsp.DocumentSymbolParams) (result []interface{}, err error)

	// DidOpen is called when a document is opened. This function has to add the document to the document store
	DidOpen(ctx context.Context, params *lsp.DidOpenTextDocumentParams, helmlsConfig util.HelmlsConfiguration) (err error)
	// DidSave is called when a document is saved, it must not update the document store
	DidSave(ctx context.Context, params *lsp.DidSaveTextDocumentParams) (err error)
	// PostDidChange is called when a document is changed, it must not update the document store
	PostDidChange(ctx context.Context, params *lsp.DidChangeTextDocumentParams) (err error)

	Configure(ctx context.Context, helmlsConfig util.HelmlsConfiguration)
	// Should return the diagnostics for the given document, this function is called after the document was opened or saved
	GetDiagnostics(uri lsp.DocumentURI) []lsp.PublishDiagnosticsParams

	// SetChartStore is called once the chart store has been initialized
	SetChartStore(chartStore *charts.ChartStore)
	// SetChartStore is called once the client has been initialized
	SetClient(client protocol.Client)
}

func (h *ServerHandler) selectLangHandler(_ context.Context, uri uri.URI) (LangHandler, error) {
	documentType, ok := h.documents.GetDocumentType(uri)

	if !ok {
		return nil, fmt.Errorf("document %s not found or has invalid language", uri)
	}

	langHandler, ok := h.langHandlers[documentType]

	if !ok {
		return nil, fmt.Errorf("language %s not supported", documentType)
	}

	return langHandler, nil
}
