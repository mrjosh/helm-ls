package yamlls

import (
	"context"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

// CustomNewClient returns the context in which Client is embedded, jsonrpc2.Conn, and the Server.
func (yamllsConnector Connector) CustomNewClient(ctx context.Context, client protocol.Client, stream jsonrpc2.Stream, logger *zap.Logger) (context.Context, jsonrpc2.Conn, protocol.Server) {
	ctx = protocol.WithClient(ctx, client)

	conn := jsonrpc2.NewConn(stream)
	conn.Go(ctx,
		protocol.Handlers(
			protocol.ClientHandler(client, yamllsConnector.customHandler),
		),
	)
	server := protocol.ServerDispatcher(conn, logger.Named("server"))

	return ctx, conn, server
}
