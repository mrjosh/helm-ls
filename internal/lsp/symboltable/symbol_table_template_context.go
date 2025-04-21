package symboltable

import (
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
)

type TemplateContextVisitor struct {
	currentContext TemplateContext
	stashedContext []TemplateContext
	symbolTable    *SymbolTable
	content        []byte
}

func NewTemplateContextVisitor(symbolTable *SymbolTable, content []byte) *TemplateContextVisitor {
	return &TemplateContextVisitor{
		currentContext: TemplateContext{},
		stashedContext: []TemplateContext{},
		symbolTable:    symbolTable,
		content:        content,
	}
}

func (v *TemplateContextVisitor) PushContext(context string) {
	v.currentContext = append(v.currentContext, context)
}

func (v *TemplateContextVisitor) PushContextMany(context []string) {
	v.currentContext = append(v.currentContext, context...)
}

func (v *TemplateContextVisitor) PopContext() {
	if len(v.currentContext) == 0 {
		return
	}
	v.currentContext = v.currentContext[:len(v.currentContext)-1]
}

func (v *TemplateContextVisitor) PopContextN(n int) {
	v.currentContext = v.currentContext[:len(v.currentContext)-n]
}

func (v *TemplateContextVisitor) StashContext() {
	v.stashedContext = append(v.stashedContext, v.currentContext)
	v.currentContext = []string{}
}

func (v *TemplateContextVisitor) RestoreStashedContext() {
	v.currentContext = v.stashedContext[len(v.stashedContext)-1]
	v.stashedContext = v.stashedContext[:len(v.stashedContext)-1]
}

func (v *TemplateContextVisitor) Enter(node *sitter.Node) {
	nodeType := node.Type()
	switch nodeType {
	case gotemplate.NodeTypeDot:
		v.symbolTable.AddTemplateContext(v.currentContext, node.Range())
	case gotemplate.NodeTypeFieldIdentifier:
		content := node.Content(v.content)
		v.symbolTable.AddTemplateContext(append(v.currentContext, content), node.Range())
	case gotemplate.NodeTypeField:
		content := node.ChildByFieldName("name").Content(v.content)
		v.symbolTable.AddTemplateContext(append(v.currentContext, content), node.ChildByFieldName("name").Range())
	case gotemplate.NodeTypeUnfinishedSelectorExpression:
		operandNode := node.ChildByFieldName("operand")
		content := getContextForSelectorExpression(operandNode, v.content)
		if !content.IsVariable() {
			content = append(v.currentContext, content...)
		}
		v.symbolTable.AddTemplateContext(append(content, ""),
			node.Child(int(node.ChildCount())-1).Range())
	case gotemplate.NodeTypeSelectorExpression:
		operandNode := node.ChildByFieldName("operand")
		if operandNode.Type() == gotemplate.NodeTypeVariable {
			v.StashContext()
			v.PushContext(operandNode.Content(v.content))
		}
	}
}

func (v *TemplateContextVisitor) Exit(node *sitter.Node) {
	switch node.Type() {
	case gotemplate.NodeTypeSelectorExpression:
		operandNode := node.ChildByFieldName("operand")
		if operandNode.Type() == gotemplate.NodeTypeVariable {
			v.PopContext()
			v.RestoreStashedContext()
		}
	}
}

func (v *TemplateContextVisitor) EnterContextShift(node *sitter.Node, suffix string) {
	switch node.Type() {
	case gotemplate.NodeTypeFieldIdentifier:
		content := node.Content(v.content) + suffix
		v.PushContext(content)
	case gotemplate.NodeTypeField:
		content := node.ChildByFieldName("name").Content(v.content) + suffix
		v.PushContext(content)
	case gotemplate.NodeTypeSelectorExpression, gotemplate.NodeTypeUnfinishedSelectorExpression:
		s := getContextForSelectorExpression(node, v.content)
		if len(s) > 0 {
			s = s.AppendSuffix(suffix)
			if s.IsVariable() {
				v.StashContext()
			}
		}
		v.PushContextMany(s)
	}
}

func (v *TemplateContextVisitor) ExitContextShift(node *sitter.Node) {
	switch node.Type() {
	case gotemplate.NodeTypeField, gotemplate.NodeTypeFieldIdentifier:
		v.PopContext()
	case gotemplate.NodeTypeSelectorExpression, gotemplate.NodeTypeUnfinishedSelectorExpression:
		s := getContextForSelectorExpression(node, v.content)
		if s.IsVariable() {
			v.RestoreStashedContext()
		} else {
			v.PopContextN(len(s))
		}
	}
}

func getContextForSelectorExpression(node *sitter.Node, content []byte) TemplateContext {
	if node == nil {
		return TemplateContext{}
	}
	if node.Type() == gotemplate.NodeTypeField {
		return TemplateContext{node.ChildByFieldName("name").Content(content)}
	}
	if node.Type() == gotemplate.NodeTypeVariable {
		return TemplateContext{node.Content(content)}
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
