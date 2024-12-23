package lsp

import (
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
)

func getTextNodeRanges(gotemplateNode *sitter.Node) []sitter.Range {
	textNodes := []sitter.Range{}

	for i := 0; i < int(gotemplateNode.ChildCount()); i++ {
		child := gotemplateNode.Child(i)
		if child.Type() == gotemplate.NodeTypeText {
			textNodes = append(textNodes, util.GetRangeForNode(child))
		} else {
			textNodes = append(textNodes, getTextNodeRanges(child)...)
		}
	}
	return textNodes
}

// TrimTemplate removes all template nodes.
// This is done by keeping only the text nodes
// which is easier then removing the template nodes
// since template nodes could contain other nodes
func TrimTemplate(gotemplateTree *sitter.Tree, content []byte) string {
	ranges := getTextNodeRanges(gotemplateTree.RootNode())
	result := make([]byte, len(content))
	for i := range result {
		if content[i] == '\n' || content[i] == '\r' {
			result[i] = content[i]
			continue
		}
		result[i] = byte(' ')
	}
	for _, yamlRange := range ranges {
		copy(result[yamlRange.StartByte:yamlRange.EndByte],
			content[yamlRange.StartByte:yamlRange.EndByte])
	}
	return string(result)
}
