package yamlls

import (
	"context"

	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector Connector) CallCompletion(ctx context.Context, params *lsp.CompletionParams) (*lsp.CompletionList, error) {
	if yamllsConnector.server == nil {
		return &lsp.CompletionList{}, nil
	}

	return yamllsConnector.server.Completion(ctx, params)
}
