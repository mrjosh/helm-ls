package lsp

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/yaml"
)

func getTextNodeRanges(gotemplateNode *sitter.Node) []sitter.Range {
	textNodes := []sitter.Range{}

	for i := 0; i < int(gotemplateNode.ChildCount()); i++ {
		child := gotemplateNode.Child(i)
		if child.Type() == gotemplate.NodeTypeText {
			textNodes = append(textNodes, sitter.Range{
				StartPoint: child.StartPoint(),
				EndPoint:   child.EndPoint(),
				StartByte:  child.StartByte(),
				EndByte:    child.EndByte(),
			})
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

func TrimTemplate(gotemplateTree *sitter.Tree, content string) string {
	ranges := getTextNodeRanges(gotemplateTree.RootNode())
	result := make([]byte, len(content))
	for i := 0; i < len(result); i++ {
		if content[i] == '\n' {
			result[i] = byte('\n')
			continue
		}
		result[i] = byte(' ')
	}
	for _, yamlRange := range ranges {
		copy(result[yamlRange.StartByte:yamlRange.EndByte], content[yamlRange.StartByte:yamlRange.EndByte])
	}
	return string(result)
}
