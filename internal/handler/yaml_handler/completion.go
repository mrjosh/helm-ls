package yamlhandler

import (
	"context"

	"go.lsp.dev/protocol"
)

// Completion implements handler.LangHandler.
func (h *YamlHandler) Completion(ctx context.Context, params *protocol.CompletionParams) (result *protocol.CompletionList, err error) {
	return h.yamllsConnector.CallCompletion(ctx, params)
}
