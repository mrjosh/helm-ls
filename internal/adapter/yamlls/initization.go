package yamlls

import (
	"context"
	"os"

	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (yamllsConnector Connector) CallInitialize(workspaceURI uri.URI) error {
	if yamllsConnector.server == nil {
		return nil
	}

	params := lsp.InitializeParams{
		RootURI:   workspaceURI,
		ProcessID: int32(os.Getpid()),
		ClientInfo: &lsp.ClientInfo{
			Name: "helm-ls",
		},
	}

	_, err := yamllsConnector.server.Initialize(context.Background(), &params)
	if err != nil {
		return err
	}
	err = yamllsConnector.server.DidChangeConfiguration(context.Background(), &lsp.DidChangeConfigurationParams{})
	if err != nil {
		return err
	}
	return yamllsConnector.server.Initialized(context.Background(), &lsp.InitializedParams{})
}
