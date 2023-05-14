package yamlls

import (
	"context"
	"os"

	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector YamllsConnector) CallInitialize(params lsp.InitializeParams) {
	if yamllsConnector.Conn == nil {
		return
	}

	params.ProcessID = int32(os.Getpid())
	params.ClientInfo.Name = "helm-ls"

	var response interface{}
	yamllsConnector.Conn.Call(context.Background(), lsp.MethodInitialize, params, response)
	yamllsConnector.Conn.Notify(context.Background(), lsp.MethodInitialized, params)

	changeConfigurationParams := lsp.DidChangeConfigurationParams{}

	yamllsConnector.Conn.Notify(context.Background(), lsp.MethodWorkspaceDidChangeConfiguration, changeConfigurationParams)
}

type YamllsSettings struct {
	Schemas    map[string]string `json:"schemas"`
	Completion bool              `json:"completion"`
	Hover      bool              `json:"hover"`
}
