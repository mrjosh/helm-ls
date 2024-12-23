package yamlhandler

import (
	"context"
	"fmt"

	"github.com/mrjosh/helm-ls/internal/protocol"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

// Hover implements handler.LangHandler.
func (h *YamlHandler) Hover(ctx context.Context, params *lsp.HoverParams) (result *lsp.Hover, err error) {
	doc, ok := h.documents.GetYamlDoc(params.TextDocument.URI)

	if !ok {
		return nil, fmt.Errorf("no document for %s", params.TextDocument.URI)
	}

	node := util.GetNodeForPosition(&doc.Node, params.Position)

	if node == nil {
		return nil, fmt.Errorf("no node found")
	}

	return protocol.BuildHoverResponse(node.Value, lsp.Range{}), doc.ParseErr
}
