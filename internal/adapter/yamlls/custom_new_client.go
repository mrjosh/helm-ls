package yamlls

import (
	"context"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

type CustomHandler struct {
	Handler        jsonrpc2.Handler
	PostInitialize SetupCustomHandler
}

type SetupCustomHandler func(ctx context.Context, conn jsonrpc2.Conn) error

var DefaultCustomHandler = CustomHandler{
	jsonrpc2.MethodNotFoundHandler, func(ctx context.Context, conn jsonrpc2.Conn) error { return nil },
}

// CustomNewClient returns the context in which Client is embedded, jsonrpc2.Conn, and the Server.
func (yamllsConnector Connector) CustomNewClient(ctx context.Context, client protocol.Client, stream jsonrpc2.Stream, logger *zap.Logger) (context.Context, jsonrpc2.Conn, protocol.Server) {
	ctx = protocol.WithClient(ctx, client)

	conn := jsonrpc2.NewConn(stream)
	conn.Go(ctx,
		protocol.Handlers(
			protocol.ClientHandler(client, yamllsConnector.customHandler.Handler),
		),
	)
	server := protocol.ServerDispatcher(conn, logger.Named("server"))

	return ctx, conn, server
}
