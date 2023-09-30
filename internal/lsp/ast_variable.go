package lsp

import (
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
)

func GetVariableDefinitionOfNode(node *sitter.Node, template string) *sitter.Node {
	if node == nil {
		return nil
	}
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

	logger.Debug("GetVariableDefinition:", node.Type(), variableName)

	switch node.Type() {
	case gotemplate.NodeTypeRangeVariableDefinition:
		var (
			indexDefinitionNode   = node.ChildByFieldName("index")
			elementDefinitionNode = node.ChildByFieldName("element")
			indexDefinitionName   = indexDefinitionNode.ChildByFieldName("name").Content([]byte(template))
			elementDefinitionName = elementDefinitionNode.ChildByFieldName("name").Content([]byte(template))
		)
		if indexDefinitionName == variableName {
			return indexDefinitionNode
		}
		if elementDefinitionName == variableName {
			return elementDefinitionNode
		}
	case gotemplate.NodeTypeVariableDefinition:
		currentVariableName := node.ChildByFieldName("variable").Child(1).Content([]byte(template))
		logger.Debug("currentVariableName:", currentVariableName)
		logger.Debug("variableName:", variableName)
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
