package handler

import (
	"errors"

	"github.com/mrjosh/helm-ls/internal/charts"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) genericDocumentUseCase(params lsp.TextDocumentPositionParams) (*lsplocal.Document, *charts.Chart, *sitter.Node, error) {
	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return nil, nil, nil, errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	chart, err := h.chartStore.GetChartForDoc(params.TextDocument.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", params.TextDocument.URI, err)
	}
	node := h.getNode(doc, params.Position)
	if node == nil {
		return doc, chart, nil, errors.New("Could not get node for: " + params.TextDocument.URI.Filename())
	}
	return doc, chart, node, nil
}

func (h *langHandler) getNode(doc *lsplocal.Document, position lsp.Position) *sitter.Node {
	var (
		currentNode   = lsplocal.NodeAtPosition(doc.Ast, position)
		pointToLookUp = sitter.Point{
			Row:    position.Line,
			Column: position.Character,
		}
	)
	return lsplocal.FindRelevantChildNode(currentNode, pointToLookUp)
}
