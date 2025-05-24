package yamlls

import (
	"context"

	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector Connector) CallCompletion(ctx context.Context, params *lsp.CompletionParams) (*lsp.CompletionList, error) {
	if !yamllsConnector.shouldRun(params.TextDocument.URI) {
		return &lsp.CompletionList{}, nil
	}

	return yamllsConnector.server.Completion(ctx, params)
}
