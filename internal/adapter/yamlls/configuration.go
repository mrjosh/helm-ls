package yamlls

import (
	"encoding/json"

	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector Connector) handleConfiguration(req jsonrpc2.Request) []interface{} {
	var params lsp.ConfigurationParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		logger.Error("Error parsing configuration request from yamlls", err)
	}
	logger.Debug("Yamlls ConfigurationParams", params)
	settings := []interface{}{yamllsConnector.config.YamllsSettings}
	return settings
}
