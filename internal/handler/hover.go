package handler

import (
	"context"

	lsp "go.lsp.dev/protocol"
)

func (h *ServerHandler) Hover(ctx context.Context, params *lsp.HoverParams) (result *lsp.Hover, err error) {
	logger.Debug("Running hover with params", params)

	handler, err := h.selectLangHandler(ctx, params.TextDocument.URI)
	if err != nil {
		logger.Error("Error selecting lang handler", err)
		return nil, err
	}
	return handler.Hover(ctx, params)
}
