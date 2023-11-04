package yamlls

import (
	"context"
	"os"

	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (yamllsConnector YamllsConnector) CallInitialize(workspaceUri uri.URI) {
	if yamllsConnector.Conn == nil {
		return
	}

	params := lsp.InitializeParams{
		RootURI:   workspaceUri,
		ProcessID: int32(os.Getpid()),
		ClientInfo: &lsp.ClientInfo{
			Name: "helm-ls",
		},
	}

	var response interface{}
	(*yamllsConnector.Conn).Call(context.Background(), lsp.MethodInitialize, params, response)
	(*yamllsConnector.Conn).Notify(context.Background(), lsp.MethodInitialized, params)

	changeConfigurationParams := lsp.DidChangeConfigurationParams{}
	(*yamllsConnector.Conn).Notify(context.Background(), lsp.MethodWorkspaceDidChangeConfiguration, changeConfigurationParams)
}
