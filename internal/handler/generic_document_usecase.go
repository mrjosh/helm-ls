package handler

import (
	"errors"

	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) NewGenericDocumentUseCase(params lsp.TextDocumentPositionParams) (languagefeatures.GenericDocumentUseCase, error) {
	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return languagefeatures.GenericDocumentUseCase{}, errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	chart, err := h.chartStore.GetChartForDoc(params.TextDocument.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", params.TextDocument.URI, err)
	}
	node := h.getNode(doc, params.Position)
	if node == nil {
		return languagefeatures.GenericDocumentUseCase{}, errors.New("Could not get node for: " + params.TextDocument.URI.Filename())
	}
	return languagefeatures.GenericDocumentUseCase{
		Document:      doc,
		DocumentStore: h.documents,
		Chart:         chart,
		Node:          node,
	}, nil
}

func (h *langHandler) getNode(doc *lsplocal.Document, position lsp.Position) *sitter.Node {
	currentNode := lsplocal.NodeAtPosition(doc.Ast, position)
	return currentNode
}
