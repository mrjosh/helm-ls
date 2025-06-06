package templatehandler

import (
	"context"

	lsp "go.lsp.dev/protocol"
)

// DocumentSymbol implements protocol.Server.
func (h *TemplateHandler) DocumentSymbol(ctx context.Context, params *lsp.DocumentSymbolParams) (result []interface{}, err error) {
	return h.yamllsConnector.CallDocumentSymbol(ctx, params)
}
