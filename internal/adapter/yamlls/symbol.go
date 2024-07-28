package yamlls

import (
	"context"

	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector Connector) CallDocumentSymbol(ctx context.Context, params *lsp.DocumentSymbolParams) (result []interface{}, err error) {
	if !yamllsConnector.shouldRun(params.TextDocument.URI) {
		return []interface{}{}, nil
	}
	return yamllsConnector.server.DocumentSymbol(ctx, params)
}
