package yamlls

import (
	"context"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector *YamllsConnector) yamllsHandler(clientConn jsonrpc2.Conn, documents *lsplocal.DocumentStore) jsonrpc2.Handler {
	return func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {

		switch req.Method() {
		case lsp.MethodTextDocumentPublishDiagnostics:
			yamllsConnector.handleDiagnostics(req, clientConn, documents)
		case lsp.MethodWorkspaceConfiguration:
			settings := yamllsConnector.handleConfiguration(req)
			return reply(ctx, settings, nil)
		}

		return reply(ctx, true, nil)
	}
}
