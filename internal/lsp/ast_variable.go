package lsp

import (
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
)

func GetVariableDefinitionOfNode(node *sitter.Node, template string) *sitter.Node {
	if node.Type() != gotemplate.NodeTypeVariable {
		return nil
	}

	variableName := node.Child(1).Content([]byte(template))

	return GetVariableDefinition(variableName, node, template)

}

func GetVariableDefinition(variableName string, node *sitter.Node, template string) *sitter.Node {
	if node == nil {
		return nil
	}

	logger.Println("GetVariableDefinition:", node.Type())

	switch node.Type() {
	case gotemplate.NodeTypeRangeVariableDefinition:
		indexDefinition := node.NamedChild(0).Child(1).Content([]byte(template))
		elementDefinition := node.NamedChild(1).Child(1).Content([]byte(template))
		if indexDefinition == variableName ||
			elementDefinition == variableName {
			return node
		}
	case gotemplate.NodeTypeVariableDefinition:
		currentVariableName := node.ChildByFieldName("variable").Child(1).Content([]byte(template))
		logger.Println("currentVariableName:", currentVariableName)
		logger.Println("variableName:", variableName)
		if currentVariableName == variableName {
			return node
		}
	}
	nextNode := node.PrevNamedSibling()
	if nextNode == nil {
		nextNode = node.Parent()
	}
	return GetVariableDefinition(variableName, nextNode, template)
}
