package yamlls

import (
	"context"
	"encoding/json"
	"errors"

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

// NewCustomSchemaHandler creates a custom schema handler
// that registers the custom schema request to yaml-language-server
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
			var requestedURIs []string
			err := json.Unmarshal(req.Params(), &requestedURIs)
			if err != nil {
				logger.Error(err)
				return reply(ctx, nil, err)
			}

			if len(requestedURIs) == 0 {
				return reply(ctx, nil, errors.New("no URI provided"))
			}

			schemaURI, err := provider(ctx, uri.New(requestedURIs[0]))
			logger.Printf("YamlHandler: custom/schema/request for uri %s returning %s", requestedURIs, schemaURI)
			if err != nil {
				return reply(ctx, nil, err)
			}
			return reply(ctx, schemaURI, nil)
		}

		return jsonrpc2.MethodNotFoundHandler(ctx, reply, req)
	}
}

// CustomNewClient returns the context in which Client is embedded, jsonrpc2.Conn, and the Server.
// This is a patched version of protocol.NewClient (see https://github.com/go-language-server/protocol/issues/53)
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
