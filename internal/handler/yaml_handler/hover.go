package yamlhandler

import (
	"context"
	"errors"

	"github.com/mrjosh/helm-ls/internal/protocol"
	lsp "go.lsp.dev/protocol"
)

// Hover implements handler.LangHandler.
func (h *YamlHandler) Hover(ctx context.Context, params *lsp.HoverParams) (result *lsp.Hover, err error) {
	logger.Debug("YamlHandler Hover", params)

	yamlResult, yamllsErr := h.yamllsConnector.CallHover(ctx, *params)
	path, err := h.getYamlPath(params.TextDocument.URI, params.Position)

	if yamlResult == nil {
		return protocol.BuildHoverResponse(path, lsp.Range{}), errors.Join(yamllsErr, err)
	}

	yamlResult.Contents.Value = yamlResult.Contents.Value + "\n\n" + path

	return yamlResult, errors.Join(yamllsErr, err)
}
