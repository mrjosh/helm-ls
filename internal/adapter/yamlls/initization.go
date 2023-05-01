package yamlls

import (
	"context"
	"os"

	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector YamllsConnector) CallInitialize(params lsp.InitializeParams) {
	params.ProcessID = int32(os.Getpid())
	params.ClientInfo.Name = "helm-ls"
	// json, _ := json.Marshal(params)

	logger.Println("Init ")
	// logger.Println("Init with", string(json))
	// logger.Println("Init with", params.InitializationOptions)

	var response interface{}
	yamllsConnector.Conn.Call(context.Background(), lsp.MethodInitialize, params, response)
	yamllsConnector.Conn.Notify(context.Background(), lsp.MethodInitialized, params)
	logger.Println("Init done ")

	changeConfigurationParams := lsp.DidChangeConfigurationParams{
		Settings: initializationOptions{Yaml: YamllsSettings{Schemas: map[string]string{"kubernetes": "**"}}}}

	logger.Println("change config", changeConfigurationParams)
	yamllsConnector.Conn.Notify(context.Background(), lsp.MethodWorkspaceDidChangeConfiguration, changeConfigurationParams)

	logger.Println("change config done")

}

type initializationOptions struct {
	Yaml YamllsSettings `json:"yaml"`
}

type YamllsSettings struct {
	Schemas map[string]string `json:"schemas"`
}
