package handler

import (
	"context"
	"encoding/json"

	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

func (h *ServerHandler) DidChangeConfiguration(ctx context.Context, params *lsp.DidChangeConfigurationParams) (err error) {
	logger.Printf("Running DidChangeConfiguration with settings %+v \n", params.Settings)

	// since the push based approach is deprecated, we always refresh the configuration
	h.retrieveWorkspaceConfiguration(ctx)

	return nil
}

func (h *ServerHandler) registerDidChangeConfiguration(ctx context.Context) {
	// For DidChangeConfiguration to be called on changes in the configuration we need to register
	h.client.RegisterCapability(ctx, &lsp.RegistrationParams{
		Registrations: []lsp.Registration{
			{
				ID:     "helm-ls",
				Method: "workspace/didChangeConfiguration",
			},
		},
	})
}

func (h *ServerHandler) retrieveWorkspaceConfiguration(ctx context.Context) {
	logger.Debug("Calling workspace/configuration")
	configurationParams := lsp.ConfigurationParams{
		Items: []lsp.ConfigurationItem{{Section: "helm-ls"}},
	}

	rawResult, err := h.client.Configuration(ctx, &configurationParams)
	if err != nil {
		logger.Println("Error calling workspace/configuration", err)
		h.initializationWithConfig(ctx)
		return
	}

	h.helmlsConfig = parseWorkspaceConfiguration(rawResult, h.helmlsConfig)
	h.helmlsConfig.YamllsConfiguration.CompileEnabledForFilesGlobObject()
	logger.Println("Workspace configuration:", h.helmlsConfig)
	h.initializationWithConfig(ctx)
}

func parseWorkspaceConfiguration(rawResult []any, currentConfig util.HelmlsConfiguration) (result util.HelmlsConfiguration) {
	logger.Debug("Raw Workspace configuration:", rawResult)

	if len(rawResult) == 0 {
		logger.Println("Workspace configuration is empty")
		return currentConfig
	}

	jsonResult, err := json.Marshal(rawResult[0])
	if err != nil {
		logger.Println("Error marshalling workspace/configuration", err)
		return currentConfig
	}

	result = currentConfig
	err = json.Unmarshal(jsonResult, &result)
	if err != nil {
		logger.Println("Error unmarshalling workspace/configuration", err)
		return currentConfig
	}
	return result
}
