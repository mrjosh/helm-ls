package lsp

import (
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
)

func IsInElseBranch(node *sitter.Node) bool {
	parent := node.Parent()

	if parent == nil {
		return false
	}

	if parent.Type() == gotemplate.NodeTypeIfAction {
		curser := sitter.NewTreeCursor(parent)
		curser.GoToFirstChild()
		for curser.GoToNextSibling() {
			fieldName := curser.CurrentFieldName()
			if fieldName == gotemplate.FieldNameAlternative || fieldName == gotemplate.FieldNameOption {
				if curser.CurrentNode().Equal(node) {
					return true
				}
			}
		}
		curser.Close()
	}
	return IsInElseBranch(parent)
}
