package handler

import (
	"context"

	lsp "go.lsp.dev/protocol"
)

type LangHandler interface {
	Completion(ctx context.Context, params *lsp.CompletionParams) (result *lsp.CompletionList, err error)
	References(ctx context.Context, params *lsp.ReferenceParams) (result []lsp.Location, err error)
}

func (h *ServerHandler) selectLangHandler(ctx context.Context, params *lsp.TextDocumentPositionParams) (LangHandler, error) {
	langID := h.documents.GetLanguageID(params.TextDocument.URI)

	return h.langHandlers[langID], nil
}
