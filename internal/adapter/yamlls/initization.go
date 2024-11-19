package yamlls

import (
	"context"
	"os"

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

	_, err := yamllsConnector.server.Initialize(ctx, &params)
	if err != nil {
		return err
	}
	err = yamllsConnector.server.DidChangeConfiguration(ctx, &lsp.DidChangeConfigurationParams{})
	if err != nil {
		return err
	}

	defer func() {
		yamllsConnector.customHandler.PostInitialize(ctx, yamllsConnector.conn)
	}()

	return yamllsConnector.server.Initialized(ctx, &lsp.InitializedParams{})
}
