package yamlls

import (
	"context"

	"go.lsp.dev/protocol"
)

// Configuration implements protocol.Client.
func (y Connector) Configuration(_ context.Context, _ *protocol.ConfigurationParams) (result []interface{}, err error) {
	settings := []interface{}{y.config.YamllsSettings}
	return settings, nil
}

func (y Connector) DidChangeConfiguration(ctx context.Context) (err error) {
	if y.server == nil {
		return nil
	}
	return y.server.DidChangeConfiguration(ctx, &protocol.DidChangeConfigurationParams{})
}
