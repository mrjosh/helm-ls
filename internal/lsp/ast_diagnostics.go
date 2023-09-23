package lsp

import sitter "github.com/smacker/go-tree-sitter"

func IsInElseBranch(node *sitter.Node) bool {

	parent := node.Parent()

	if parent == nil {
		return false
	}

	if parent.Type() == "if_action" {

		childIndex, err := getIndexOfChild(parent, node)
		if err != nil {
			return IsInElseBranch(parent)
		}

		logger.Println("ChildIndex is", childIndex)

		if childIndex > 4 {
			return true
		}

	}

	return IsInElseBranch(parent)
}

func getIndexOfChild(parent *sitter.Node, child *sitter.Node) (int, error) {

	count := parent.ChildCount()
	for i := 0; i < int(count); i++ {
		if parent.Child(i) == child {
			return i, nil
		}
	}
	return -1, nil
}
