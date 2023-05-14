package yamlls

import (
	"encoding/json"

	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
)

func handleConfiguration(req jsonrpc2.Request) [5]interface{} {
	var params lsp.ConfigurationParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		logger.Println("Error ")
	}
	settings := [5]interface{}{YamllsSettings{Schemas: map[string]string{"kubernetes": "**"}, Completion: true, Hover: true}}
	return settings
}
