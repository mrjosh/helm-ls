package lsp

import sitter "github.com/smacker/go-tree-sitter"

func IsInElseBranch(node *sitter.Node) bool {
	parent := node.Parent()

	if parent == nil {
		return false
	}

	if parent.Type() == "if_action" {
		curser := sitter.NewTreeCursor(parent)
		curser.GoToFirstChild()
		for curser.GoToNextSibling() {
			fieldName := curser.CurrentFieldName()
			if fieldName == "alternative" || fieldName == "option" {
				if curser.CurrentNode().Equal(node) {
					return true
				}
			}
		}
		curser.Close()
	}
	return IsInElseBranch(parent)
}
