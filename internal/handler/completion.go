package handler

import (
	"context"

	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/protocol"
	lsp "go.lsp.dev/protocol"

	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
)

func (h *ServerHandler) Completion(ctx context.Context, params *lsp.CompletionParams) (result *lsp.CompletionList, err error) {
	logger.Debug("Running completion with params", params)

	return h.selectLangHandler(ctx, params).Completion(ctx, params)
}
