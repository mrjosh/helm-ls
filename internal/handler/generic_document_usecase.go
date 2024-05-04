package handler

import (
	"errors"

	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) NewGenericDocumentUseCase(
	params lsp.TextDocumentPositionParams,
	nodeSelection func(ast *sitter.Tree, position lsp.Position) (node *sitter.Node),
) (*languagefeatures.GenericDocumentUseCase, error) {
	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return &languagefeatures.GenericDocumentUseCase{}, errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	chart, err := h.chartStore.GetChartForDoc(params.TextDocument.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", params.TextDocument.URI, err)
	}

	node := nodeSelection(doc.Ast, params.Position)
	if node == nil {
		return &languagefeatures.GenericDocumentUseCase{}, errors.New("Could not get node for: " + params.TextDocument.URI.Filename())
	}

	var (
		nodeType       = node.Type()
		parentNode     = node.Parent()
		parentNodeType string
	)
	if parentNode != nil {
		parentNodeType = parentNode.Type()
	}

	return &languagefeatures.GenericDocumentUseCase{
		Document:       doc,
		DocumentStore:  h.documents,
		Chart:          chart,
		ChartStore:     h.chartStore,
		Node:           node,
		NodeType:       nodeType,
		ParentNode:     parentNode,
		ParentNodeType: parentNodeType,
	}, nil
}
