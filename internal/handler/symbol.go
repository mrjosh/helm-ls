package handler

import (
	"context"

	lsp "go.lsp.dev/protocol"
)

// DocumentSymbol implements protocol.Server.
func (h *ServerHandler) DocumentSymbol(ctx context.Context, params *lsp.DocumentSymbolParams) (result []interface{}, err error) {
	logger.Debug("Running DocumentSymbol with params", params)

	handler, err := h.selectLangHandler(ctx, params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	return handler.DocumentSymbol(ctx, params)
}
