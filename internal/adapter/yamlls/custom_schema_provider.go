package yamlls

import (
	"context"
	"encoding/json"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
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

func NewCustomSchemaHandler(handler jsonrpc2.Handler) *CustomHandler {
	return &CustomHandler{
		Handler: handler,
		PostInitialize: func(ctx context.Context, conn jsonrpc2.Conn) error {
			return conn.Notify(ctx, "yaml/registerCustomSchemaRequest", nil)
		},
	}
}

type CustomSchemaProvider func(ctx context.Context, uri uri.URI) (uri.URI, error)

func NewCustomSchemaProviderHandler(provider CustomSchemaProvider) jsonrpc2.Handler {
	return func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		switch req.Method() {
		case "custom/schema/request":

			params := []string{}
			jsonBytes, err := req.Params().MarshalJSON()
			if err != nil {
				logger.Error(err)
				return reply(ctx, nil, nil)
			}

			err = json.Unmarshal(jsonBytes, &params)
			if err != nil {
				logger.Error(err)
				return reply(ctx, nil, nil)
			}

			logger.Println("YamlHandler: custom/schema/request", string(req.Params()))

			if len(params) == 0 {
				return reply(ctx, nil, nil)
			}

			schemaURI, err := provider(ctx, uri.New(params[0]))
			if err != nil {
				return reply(ctx, nil, err)
			}
			return reply(ctx, schemaURI, nil)
		}
		return jsonrpc2.MethodNotFoundHandler(ctx, reply, req)
	}
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
