package lsp

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/yaml"
)

func getRangeForNode(node *sitter.Node) sitter.Range {
	return sitter.Range{
		StartPoint: node.StartPoint(),
		EndPoint:   node.EndPoint(),
		StartByte:  node.StartByte(),
		EndByte:    node.EndByte(),
	}
}

func getTextNodeRanges(gotemplateNode *sitter.Node) []sitter.Range {
	textNodes := []sitter.Range{}

	for i := 0; i < int(gotemplateNode.ChildCount()); i++ {
		child := gotemplateNode.Child(i)
		if child.Type() == gotemplate.NodeTypeText {
			textNodes = append(textNodes, getRangeForNode(child))
		} else {
			textNodes = append(textNodes, getTextNodeRanges(child)...)
		}
	}
	return textNodes
}

func ParseYamlAst(gotemplateTree *sitter.Tree, content string) *sitter.Tree {
	parser := sitter.NewParser()
	parser.SetLanguage(yaml.GetLanguage())
	parser.SetIncludedRanges(getTextNodeRanges(gotemplateTree.RootNode()))

	tree, _ := parser.ParseCtx(context.Background(), nil, []byte(content))
	return tree
}

// TrimTemplate removes all template nodes.
// This is done by keeping only the text nodes
// which is easier then removing the template nodes
// since template nodes could contain other nodes
func TrimTemplate(gotemplateTree *sitter.Tree, content string) string {
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
