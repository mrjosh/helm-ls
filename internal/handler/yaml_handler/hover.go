package yamlhandler

import (
	"context"

	lsp "go.lsp.dev/protocol"
)

// Hover implements handler.LangHandler.
func (h *YamlHandler) Hover(ctx context.Context, params *lsp.HoverParams) (result *lsp.Hover, err error) {
	logger.Debug("YamlHandler Hover", params)
	return h.yamllsConnector.CallHover(ctx, *params)
	// doc, ok := h.documents.GetYamlDoc(params.TextDocument.URI)
	//
	// if !ok {
	// 	return nil, fmt.Errorf("no document for %s", params.TextDocument.URI)
	// }
	//
	// node := util.GetNodeForPosition(&doc.Node, params.Position)
	//
	// if node == nil {
	// 	return nil, fmt.Errorf("no node found")
	// }
	//
	// return protocol.BuildHoverResponse(node.Value, lsp.Range{}), doc.ParseErr
}
