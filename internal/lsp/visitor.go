package lsp

import (
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
)

func (v *Visitors) parseNodesRecursive(node *sitter.Node) {
	for _, visitor := range v.visitors {
		visitor.Enter(node)
	}
	for i := uint32(0); i < node.ChildCount(); i++ {
		v.parseNodesRecursive(node.Child(int(i)))
	}
	for _, visitor := range v.visitors {
		visitor.Exit(node)
	}
}

type Visitors struct {
	visitors    []Visitor
	symbolTable *SymbolTable
}

type Visitor interface {
	Enter(node *sitter.Node)
	Exit(node *sitter.Node)
	EnterScopeShift(node *sitter.Node, suffix string)
	ExitScopeShift(node *sitter.Node)
}

func (v *Visitors) visitNodesRecursiveWithScopeShift(node *sitter.Node) {
	for _, visitor := range v.visitors {
		visitor.Enter(node)
	}

	nodeType := node.Type()
	switch nodeType {
	case gotemplate.NodeTypeWithAction:
		condition := node.ChildByFieldName("condition")
		v.visitNodesRecursiveWithScopeShift(condition)
		for _, visitor := range v.visitors {
			visitor.EnterScopeShift(condition, "")
		}
		for i := uint32(1); i < node.NamedChildCount(); i++ {
			consequence := node.NamedChild(int(i))
			v.visitNodesRecursiveWithScopeShift(consequence)
		}
		for _, visitor := range v.visitors {
			visitor.ExitScopeShift(condition)
		}
	case gotemplate.NodeTypeRangeAction:
		rangeNode := node.ChildByFieldName("range")
		v.visitNodesRecursiveWithScopeShift(rangeNode)
		for _, visitor := range v.visitors {
			visitor.EnterScopeShift(rangeNode, "[]")
		}
		for i := uint32(1); i < node.NamedChildCount(); i++ {
			body := node.NamedChild(int(i))
			v.visitNodesRecursiveWithScopeShift(body)
		}
		for _, visitor := range v.visitors {
			visitor.ExitScopeShift(rangeNode)
		}
	case gotemplate.NodeTypeSelectorExpression:
		operand := node.ChildByFieldName("operand")
		v.visitNodesRecursiveWithScopeShift(operand)
		for _, visitor := range v.visitors {
			visitor.EnterScopeShift(operand, "")
		}
		field := node.ChildByFieldName("field")
		v.visitNodesRecursiveWithScopeShift(field)
		for _, visitor := range v.visitors {
			visitor.ExitScopeShift(operand)
		}

	default:
		for i := uint32(0); i < node.ChildCount(); i++ {
			v.visitNodesRecursiveWithScopeShift(node.Child(int(i)))
		}
	}

	for _, visitor := range v.visitors {
		visitor.Exit(node)
	}
}
