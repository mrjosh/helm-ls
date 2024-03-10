package yamlls

import (
	"context"
	"os"

	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (yamllsConnector Connector) CallInitialize(workspaceURI uri.URI) (result *lsp.InitializeResult, err error) {
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

	return yamllsConnector.server.Initialize(context.Background(), &params)
}
