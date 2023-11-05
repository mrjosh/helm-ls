package handler

import (
	"context"
	// "reflect"

	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) handleWorkspaceDidChangeConfiguration(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {
	go h.retrieveWorkspaceConfiguration(ctx)
	return reply(ctx, nil, nil)
}

func (h *langHandler) retrieveWorkspaceConfiguration(ctx context.Context) {
	logger.Println("Calling workspace/configuration")
	result := []util.HelmlsConfiguration{util.DefaultConfig}

	_, err := h.connPool.Call(ctx, lsp.MethodWorkspaceConfiguration, lsp.ConfigurationParams{
		Items: []lsp.ConfigurationItem{{Section: "helm-ls"}},
	}, &result)

	if err != nil {
		logger.Println("Error calling workspace/configuration", err)
	} else {
		logger.Println("Workspace configuration:", result)
	}

	h.helmlsConfig = result[0]
	h.initializationWithConfig(ctx)
}
