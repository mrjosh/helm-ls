package lsp

import (
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
)

type ValuesVisitor struct {
	currentContext []string
	stashedContext [][]string
	symbolTable    *SymbolTable
	content        []byte
}

func (v *ValuesVisitor) Enter(node *sitter.Node) {
	switch node.Type() {
	case gotemplate.NodeTypeDot:
		v.symbolTable.AddValue(v.currentContext, getRangeForNode(node))
	case gotemplate.NodeTypeFieldIdentifier:
		content := node.Content(v.content)
		v.symbolTable.AddValue(append(v.currentContext, content), getRangeForNode(node))
	case gotemplate.NodeTypeField:
		content := node.ChildByFieldName("name").Content(v.content)
		v.symbolTable.AddValue(append(v.currentContext, content), getRangeForNode(node))
	case gotemplate.NodeTypeSelectorExpression:
		operandNode := node.ChildByFieldName("operand")
		if operandNode.Type() == gotemplate.NodeTypeVariable && operandNode.Content(v.content) == "$" {
			v.stashedContext = append(v.stashedContext, v.currentContext)
			v.currentContext = []string{}
		}
	}
}

func (v *ValuesVisitor) Exit(node *sitter.Node) {
	switch node.Type() {
	case gotemplate.NodeTypeSelectorExpression:
		operandNode := node.ChildByFieldName("operand")
		if operandNode.Type() == gotemplate.NodeTypeVariable && operandNode.Content(v.content) == "$" {
			v.currentContext = v.stashedContext[len(v.stashedContext)-1]
			v.stashedContext = v.stashedContext[:len(v.stashedContext)-1]
		}
	}
}

func (v *ValuesVisitor) EnterContextShift(node *sitter.Node, suffix string) {
	switch node.Type() {
	case gotemplate.NodeTypeFieldIdentifier:
		content := node.Content(v.content) + suffix
		v.currentContext = append(v.currentContext, content)
	case gotemplate.NodeTypeField:
		content := node.ChildByFieldName("name").Content(v.content) + suffix
		v.currentContext = append(v.currentContext, content)
	case gotemplate.NodeTypeSelectorExpression:
		s := getContextForSelectorExpression(node, v.content)
		if len(s) > 0 {
			s[len(s)-1] = s[len(s)-1] + suffix
			if s[0] == "$" {
				v.stashedContext = append(v.stashedContext, v.currentContext)
				v.currentContext = []string{}
				s = s[1:]
			}
		}
		v.currentContext = append(v.currentContext, s...)
	}
}

func (v *ValuesVisitor) ExitContextShift(node *sitter.Node) {
	switch node.Type() {
	case gotemplate.NodeTypeField, gotemplate.NodeTypeFieldIdentifier:
		v.currentContext = v.currentContext[:len(v.currentContext)-1]
	case gotemplate.NodeTypeSelectorExpression:
		s := getContextForSelectorExpression(node, v.content)
		if len(s) > 0 && s[0] == "$" {
			v.currentContext = v.stashedContext[len(v.stashedContext)-1]
			v.stashedContext = v.stashedContext[:len(v.stashedContext)-1]
			s = s[1:]
		} else {
			v.currentContext = v.currentContext[:len(v.currentContext)-len(s)]
		}
	}
}

func getContextForSelectorExpression(node *sitter.Node, content []byte) []string {
	if node == nil {
		return []string{}
	}
	if node.Type() == gotemplate.NodeTypeField {
		return []string{node.ChildByFieldName("name").Content(content)}
	}
	if node.Type() == gotemplate.NodeTypeVariable {
		return []string{node.Content(content)}
	}

	operand := node.ChildByFieldName("operand")
	operandScope := getContextForSelectorExpression(operand, content)
	field := node.ChildByFieldName("field")
	if field == nil {
		return operandScope
	}
	fieldScope := field.Content(content)

	return append(operandScope, fieldScope)
}
