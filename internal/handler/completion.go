package handler

import (
	"context"

	lsp "go.lsp.dev/protocol"
)

func (h *ServerHandler) Completion(ctx context.Context, params *lsp.CompletionParams) (result *lsp.CompletionList, err error) {
	logger.Debug("Running completion with params", params)

	handler, err := h.selectLangHandler(ctx, params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	return handler.Completion(ctx, params)
}
