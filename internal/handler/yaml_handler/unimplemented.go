package yamlhandler

import (
	"context"

	"go.lsp.dev/protocol"
)

// DocumentSymbol implements handler.LangHandler.
func (h *YamlHandler) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) (result []interface{}, err error) {
	return nil, nil
}
