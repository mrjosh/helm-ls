package yamlls

import (
	"context"
	"fmt"

	"go.lsp.dev/protocol"
)

// ApplyEdit implements protocol.Client.
func (y Connector) ApplyEdit(ctx context.Context, params *protocol.ApplyWorkspaceEditParams) (result bool, err error) {
	return true, nil
}

// LogMessage implements protocol.Client.
func (y Connector) LogMessage(ctx context.Context, params *protocol.LogMessageParams) (err error) {
	logger.Debug(fmt.Sprintf("LogMessage from yamlls: %s %s", params.Type, params.Message))
	return nil
}

// Progress implements protocol.Client.
func (y Connector) Progress(ctx context.Context, params *protocol.ProgressParams) (err error) {
	return nil
}

// RegisterCapability implements protocol.Client.
func (y Connector) RegisterCapability(ctx context.Context, params *protocol.RegistrationParams) (err error) {
	return nil
}

// ShowMessage implements protocol.Client.
func (y Connector) ShowMessage(ctx context.Context, params *protocol.ShowMessageParams) (err error) {
	return y.client.ShowMessage(ctx, params)
}

// ShowMessageRequest implements protocol.Client.
func (y Connector) ShowMessageRequest(ctx context.Context, params *protocol.ShowMessageRequestParams) (result *protocol.MessageActionItem, err error) {
	return nil, nil
}

// Telemetry implements protocol.Client.
func (y Connector) Telemetry(ctx context.Context, params interface{}) (err error) {
	return nil
}

// UnregisterCapability implements protocol.Client.
func (y Connector) UnregisterCapability(ctx context.Context, params *protocol.UnregistrationParams) (err error) {
	return nil
}

// WorkDoneProgressCreate implements protocol.Client.
func (y Connector) WorkDoneProgressCreate(ctx context.Context, params *protocol.WorkDoneProgressCreateParams) (err error) {
	return nil
}

// WorkspaceFolders implements protocol.Client.
func (y Connector) WorkspaceFolders(ctx context.Context) (result []protocol.WorkspaceFolder, err error) {
	return nil, nil
}
