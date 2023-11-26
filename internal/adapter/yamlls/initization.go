package yamlls

import (
	"context"
	"os"

	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (yamllsConnector Connector) CallInitialize(workspaceURI uri.URI) {
	if yamllsConnector.Conn == nil {
		return
	}

	params := lsp.InitializeParams{
		RootURI:   workspaceURI,
		ProcessID: int32(os.Getpid()),
		ClientInfo: &lsp.ClientInfo{
			Name: "helm-ls",
		},
	}

	var response interface{}
	_, err := (*yamllsConnector.Conn).Call(context.Background(), lsp.MethodInitialize, params, response)
	if err != nil {
		logger.Error("Error calling yamlls for initialize", err)
		return
	}
	(*yamllsConnector.Conn).Notify(context.Background(), lsp.MethodInitialized, params)

	changeConfigurationParams := lsp.DidChangeConfigurationParams{}
	(*yamllsConnector.Conn).Notify(context.Background(), lsp.MethodWorkspaceDidChangeConfiguration, changeConfigurationParams)
}
