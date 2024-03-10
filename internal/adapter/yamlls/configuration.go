package yamlls

import (
	"context"

	"go.lsp.dev/protocol"
)

// Configuration implements protocol.Client.
func (y Connector) Configuration(ctx context.Context, params *protocol.ConfigurationParams) (result []interface{}, err error) {
	settings := []interface{}{y.config.YamllsSettings}
	return settings, nil
}

func (y Connector) DidChangeConfiguration() (err error) {
	ctx := context.Background()
	return y.server.DidChangeConfiguration(ctx, &protocol.DidChangeConfigurationParams{})
}
