package handler

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

// References implements protocol.Server.
func (h *langHandler) References(ctx context.Context, params *lsp.ReferenceParams) (result []lsp.Location, err error) {
	doc, _, node, err := h.genericDocumentUseCase(params.TextDocumentPositionParams)
	if err != nil {
		return nil, err
	}

	parentNode := node.Parent()
	pt := parentNode.Type()
	ct := node.Type()

	if pt == gotemplate.NodeTypeDefineAction && ct == gotemplate.NodeTypeInterpretedStringLiteral {
		referenceRanges, ok := doc.SymbolTable.GetIncludeReference(util.RemoveQuotes(node.Content([]byte(doc.Content))))

		locations := []lsp.Location{}
		for _, referenceRange := range referenceRanges {
			locations = append(locations, lsp.Location{
				URI: params.TextDocumentPositionParams.TextDocument.URI,
				Range: lsp.Range{
					Start: util.PointToPosition(referenceRange.StartPoint),
					End:   util.PointToPosition(referenceRange.EndPoint),
				},
			})
		}

		if ok {
			return locations, nil
		}
	}
	return nil, nil
}
