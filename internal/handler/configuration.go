package handler

import (
	"context"
	// "reflect"

	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) DidChangeConfiguration(ctx context.Context, params *lsp.DidChangeConfigurationParams) (err error) {
	// go h.retrieveWorkspaceConfiguration(ctx)
	logger.Println("Changing workspace config is not implemented")
	return nil
}

func (h *langHandler) RetrieveWorkspaceConfiguration(ctx context.Context) {
	logger.Println("Calling workspace/configuration")
	result := []util.HelmlsConfiguration{util.DefaultConfig}

	configurationParams := lsp.ConfigurationParams{
		Items: []lsp.ConfigurationItem{{Section: "helm-ls"}},
	}

	rawResult, err := h.client.Configuration(ctx, &configurationParams)

	if err != nil {
		logger.Println("Error calling workspace/configuration", err)
	} else {
		logger.Println("Workspace configuration:", rawResult)
	}

	if len(result) == 0 {
		logger.Println("Workspace configuration is empty")
		return
	}

	// TODO: use retrieved config
	h.helmlsConfig = util.DefaultConfig

	logger.Println("Workspace configuration:", h.helmlsConfig)
	h.initializationWithConfig()
}
