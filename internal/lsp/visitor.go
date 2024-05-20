package lsp

import (
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
)

type Visitors struct {
	visitors    []Visitor
	symbolTable *SymbolTable
}

type Visitor interface {
	Enter(node *sitter.Node)
	Exit(node *sitter.Node)
	EnterContextShift(node *sitter.Node, suffix string)
	ExitContextShift(node *sitter.Node)
}

func (v *Visitors) visitNodesRecursiveWithScopeShift(node *sitter.Node) {
	if node == nil {
		return
	}
	for _, visitor := range v.visitors {
		visitor.Enter(node)
	}

	nodeType := node.Type()
	switch nodeType {
	case gotemplate.NodeTypeWithAction:
		condition := node.ChildByFieldName("condition")
		v.visitNodesRecursiveWithScopeShift(condition)
		for _, visitor := range v.visitors {
			visitor.EnterContextShift(condition, "")
		}
		for i := uint32(1); i < node.NamedChildCount(); i++ {
			consequence := node.NamedChild(int(i))
			v.visitNodesRecursiveWithScopeShift(consequence)
		}
		for _, visitor := range v.visitors {
			visitor.ExitContextShift(condition)
		}
	case gotemplate.NodeTypeRangeAction:
		rangeNode := node.ChildByFieldName("range")
		if rangeNode == nil {
			// for {{- range $type, $config := $root.Values.deployments }} the range node is in the
			// range_variable_definition node an not in the range_action node
			rangeNode = node.NamedChild(0).ChildByFieldName("range")
			if rangeNode == nil {
				logger.Error("Could not find range node")
				break
			}
		}
		v.visitNodesRecursiveWithScopeShift(rangeNode)
		for _, visitor := range v.visitors {
			visitor.EnterContextShift(rangeNode, "[]")
		}
		for i := uint32(1); i < node.NamedChildCount(); i++ {
			body := node.NamedChild(int(i))
			v.visitNodesRecursiveWithScopeShift(body)
		}
		for _, visitor := range v.visitors {
			visitor.ExitContextShift(rangeNode)
		}
	case gotemplate.NodeTypeSelectorExpression, gotemplate.NodeTypeUnfinishedSelectorExpression:
		operand := node.ChildByFieldName("operand")
		v.visitNodesRecursiveWithScopeShift(operand)
		for _, visitor := range v.visitors {
			visitor.EnterContextShift(operand, "")
		}
		field := node.ChildByFieldName("field")
		v.visitNodesRecursiveWithScopeShift(field)
		for _, visitor := range v.visitors {
			visitor.ExitContextShift(operand)
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
