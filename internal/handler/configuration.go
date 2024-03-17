package handler

import (
	"context"
	"encoding/json"

	// "reflect"

	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) DidChangeConfiguration(ctx context.Context, params *lsp.DidChangeConfigurationParams) (err error) {
	// go h.retrieveWorkspaceConfiguration(ctx)
	logger.Println("Changing workspace config is not implemented")
	return nil
}

func (h *langHandler) retrieveWorkspaceConfiguration(ctx context.Context) {
	logger.Debug("Calling workspace/configuration")
	configurationParams := lsp.ConfigurationParams{
		Items: []lsp.ConfigurationItem{{Section: "helm-ls"}},
	}
	result := h.helmlsConfig

	rawResult, err := h.client.Configuration(ctx, &configurationParams)
	if err != nil {
		logger.Println("Error calling workspace/configuration", err)
		h.initializationWithConfig(ctx)
		return
	}

	logger.Debug("Raw Workspace configuration:", rawResult)

	if len(rawResult) == 0 {
		logger.Println("Workspace configuration is empty")
		h.initializationWithConfig(ctx)
		return
	}

	jsonResult, err := json.Marshal(rawResult[0])
	if err != nil {
		logger.Println("Error marshalling workspace/configuration", err)
		h.initializationWithConfig(ctx)
		return
	}
	err = json.Unmarshal(jsonResult, &result)
	if err != nil {
		logger.Println("Error unmarshalling workspace/configuration", err)
		h.initializationWithConfig(ctx)
		return
	}

	h.helmlsConfig = result

	logger.Println("Workspace configuration:", h.helmlsConfig)
	h.initializationWithConfig(ctx)
}
