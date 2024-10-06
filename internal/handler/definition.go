package handler

import (
	"context"

	lsp "go.lsp.dev/protocol"
)

func (h *ServerHandler) Definition(ctx context.Context, params *lsp.DefinitionParams) (result []lsp.Location, err error) {
	logger.Debug("Running Definition with params", params)

	handler, err := h.selectLangHandler(ctx, params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	return handler.Definition(ctx, params)
}
