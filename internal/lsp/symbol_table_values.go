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

func NewValuesVisitor(symbolTable *SymbolTable, content []byte) *ValuesVisitor {
	return &ValuesVisitor{
		currentContext: []string{},
		stashedContext: [][]string{},
		symbolTable:    symbolTable,
		content:        content,
	}
}

func (v *ValuesVisitor) PushContext(context string) {
	v.currentContext = append(v.currentContext, context)
}

func (v *ValuesVisitor) PushContextMany(context []string) {
	v.currentContext = append(v.currentContext, context...)
}

func (v *ValuesVisitor) PopContext() {
	v.currentContext = v.currentContext[:len(v.currentContext)-1]
}

func (v *ValuesVisitor) PopContextN(n int) {
	v.currentContext = v.currentContext[:len(v.currentContext)-n]
}

func (v *ValuesVisitor) StashContext() {
	v.stashedContext = append(v.stashedContext, v.currentContext)
	v.currentContext = []string{}
}

func (v *ValuesVisitor) RestoreStashedContext() {
	v.currentContext = v.stashedContext[len(v.stashedContext)-1]
	v.stashedContext = v.stashedContext[:len(v.stashedContext)-1]
}

func (v *ValuesVisitor) Enter(node *sitter.Node) {
	switch node.Type() {
	case gotemplate.NodeTypeDot:
		v.symbolTable.AddValue(v.currentContext, GetRangeForNode(node))
	case gotemplate.NodeTypeFieldIdentifier:
		content := node.Content(v.content)
		v.symbolTable.AddValue(append(v.currentContext, content), GetRangeForNode(node))
	case gotemplate.NodeTypeField:
		content := node.ChildByFieldName("name").Content(v.content)
		v.symbolTable.AddValue(append(v.currentContext, content), GetRangeForNode(node))
	case gotemplate.NodeTypeSelectorExpression:
		operandNode := node.ChildByFieldName("operand")
		if operandNode.Type() == gotemplate.NodeTypeVariable && operandNode.Content(v.content) == "$" {
			v.StashContext()
		}
	}
}

func (v *ValuesVisitor) Exit(node *sitter.Node) {
	switch node.Type() {
	case gotemplate.NodeTypeSelectorExpression:
		operandNode := node.ChildByFieldName("operand")
		if operandNode.Type() == gotemplate.NodeTypeVariable && operandNode.Content(v.content) == "$" {
			v.RestoreStashedContext()
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
		v.PushContext(content)
	case gotemplate.NodeTypeSelectorExpression:
		s := getContextForSelectorExpression(node, v.content)
		if len(s) > 0 {
			s[len(s)-1] = s[len(s)-1] + suffix
			if s[0] == "$" {
				v.StashContext()
				s = s[1:]
			}
		}
		v.PushContextMany(s)
	}
}

func (v *ValuesVisitor) ExitContextShift(node *sitter.Node) {
	switch node.Type() {
	case gotemplate.NodeTypeField, gotemplate.NodeTypeFieldIdentifier:
		v.PopContext()
	case gotemplate.NodeTypeSelectorExpression:
		s := getContextForSelectorExpression(node, v.content)
		if len(s) > 0 && s[0] == "$" {
			v.RestoreStashedContext()
		} else {
			v.PopContextN(len(s))
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
