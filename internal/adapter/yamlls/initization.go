package yamlls

import (
	"context"
	"os"
	"time"

	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (yamllsConnector Connector) CallInitialize(ctx context.Context, workspaceURI uri.URI) error {
	if yamllsConnector.server == nil {
		return nil
	}

	params := lsp.InitializeParams{
		RootURI:   workspaceURI,
		ProcessID: int32(os.Getpid()),
		ClientInfo: &lsp.ClientInfo{
			Name: "helm-ls",
		},
		Capabilities: lsp.ClientCapabilities{
			TextDocument: &lsp.TextDocumentClientCapabilities{
				DocumentSymbol: &lsp.DocumentSymbolClientCapabilities{
					HierarchicalDocumentSymbolSupport: true,
				},
			},
		},
	}

	logger.Debug("Calling yamlls initialize")

	timeout := time.Duration(yamllsConnector.config.InitTimeoutSeconds) * time.Second
	ctxT, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := yamllsConnector.server.Initialize(ctxT, &params)
	if err != nil {
		logger.Error("Error calling yamlls for initialize ", err)
		return err
	}
	logger.Debug("Calling yamlls for didChangeConfiguration")
	err = yamllsConnector.server.DidChangeConfiguration(ctx, &lsp.DidChangeConfigurationParams{})
	if err != nil {
		return err
	}

	defer func() {
		if err := yamllsConnector.customHandler.PostInitialize(ctx, yamllsConnector.conn); err != nil {
			logger.Error("Failed to post-initialize custom handler:", err)
		}
	}()

	return yamllsConnector.server.Initialized(ctx, &lsp.InitializedParams{})
}
