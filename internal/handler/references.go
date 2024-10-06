package handler

import (
	"context"

	lsp "go.lsp.dev/protocol"
)

func (h *ServerHandler) References(ctx context.Context, params *lsp.ReferenceParams) (result []lsp.Location, err error) {
	logger.Debug("Running references with params", params)

	handler, err := h.selectLangHandler(ctx, params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	return handler.References(ctx, params)
}
