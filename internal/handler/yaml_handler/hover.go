package yamlhandler

import (
	"context"
	"errors"
	"fmt"

	"github.com/mrjosh/helm-ls/internal/protocol"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

// Hover implements handler.LangHandler.
func (h *YamlHandler) Hover(ctx context.Context, params *lsp.HoverParams) (result *lsp.Hover, err error) {
	yamlResult, yamllsErr := h.yamllsConnector.CallHover(ctx, *params)

	path, err := h.GetYamlPath(params.TextDocument.URI, params.Position)

	logger.Debug("YamlHandler Hover", params)

	if yamlResult == nil {
		return protocol.BuildHoverResponse(path, lsp.Range{}), errors.Join(yamllsErr, err)
	}

	yamlResult.Contents.Value = yamlResult.Contents.Value + "\n\n" + path

	return yamlResult, errors.Join(yamllsErr, err)
}

func (h *YamlHandler) GetYamlPath(uri lsp.URI, pos lsp.Position) (path string, err error) {
	// IDEA: return the json path of the current node
	doc, ok := h.documents.GetYamlDoc(uri)

	if !ok {
		return "", fmt.Errorf("no document for %s", uri)
	}

	node := util.GetNodeForPosition2(doc.GoccyYamlNode, pos)

	if node == nil {
		return "", fmt.Errorf("no node found")
	}

	path = node.GetPath()

	return path, nil
}
