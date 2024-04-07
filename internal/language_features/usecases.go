package languagefeatures

import (
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
)

// interface for use cases
type UseCase interface {
	AppropriateForNode(currentNodeType string, parentNodeType string, node *sitter.Node) bool
}

type ReferencesUseCase interface {
	UseCase
	References() (result []lsp.Location, err error)
}
