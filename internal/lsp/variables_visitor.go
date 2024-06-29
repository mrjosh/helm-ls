package lsp

import (
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
)

type VariablesVisitor struct {
	symbolTable      *SymbolTable
	content          []byte
	insideDefinition bool
	scopeStack       []*sitter.Node
}

func NewVariablesVisitor(symbolTable *SymbolTable, content []byte) *VariablesVisitor {
	return &VariablesVisitor{
		symbolTable:      symbolTable,
		content:          content,
		insideDefinition: false,
		scopeStack:       []*sitter.Node{},
	}
}

func (v *VariablesVisitor) Enter(node *sitter.Node) {
	switch node.Type() {
	case gotemplate.NodeTypeRangeVariableDefinition:
		v.insideDefinition = true
		keyOrIndexVariableName := node.ChildByFieldName("index")
		valueVariableName := node.ChildByFieldName("element")
		variableValueNode := node.ChildByFieldName("range")
		if variableValueNode == nil {
			return
		}
		if keyOrIndexVariableName != nil {
			v.addVariableDefinition(VariableTypeRangeKeyOrIndex, node, keyOrIndexVariableName, variableValueNode)
		}
		if valueVariableName != nil {
			v.addVariableDefinition(VariableTypeRangeValue, node, valueVariableName, variableValueNode)
		}

	case gotemplate.NodeTypeVariableDefinition:
		v.insideDefinition = true
		variableNameNode := node.ChildByFieldName("variable")
		variableValueNode := node.ChildByFieldName("value")
		if variableNameNode == nil || variableValueNode == nil {
			return
		}
		v.addVariableDefinition(VariableTypeAssigment, node, variableNameNode, variableValueNode)

	case gotemplate.NodeTypeVariable:
		if v.insideDefinition {
			return
		}
		v.addVariableUsage(node)
	case gotemplate.NodeTypeIfAction,
		gotemplate.NodeTypeWithAction,
		gotemplate.NodeTypeBlockAction,
		gotemplate.NodeTypeRangeAction,
		gotemplate.NodeTypeDefineAction,
		gotemplate.NodeTypeTemplate:
		v.enterScope(node)
	}
}

func (v *VariablesVisitor) Exit(node *sitter.Node) {
	switch node.Type() {
	case gotemplate.NodeTypeRangeVariableDefinition, gotemplate.NodeTypeVariableDefinition:
		v.insideDefinition = false
	case gotemplate.NodeTypeIfAction,
		gotemplate.NodeTypeWithAction,
		gotemplate.NodeTypeBlockAction,
		gotemplate.NodeTypeRangeAction,
		gotemplate.NodeTypeDefineAction,
		gotemplate.NodeTypeTemplate:
		v.exitScope()
	}
}

func (v *VariablesVisitor) enterScope(node *sitter.Node) {
	v.scopeStack = append(v.scopeStack, node)
}

func (v *VariablesVisitor) exitScope() {
	if len(v.scopeStack) == 0 {
		return
	}
	v.scopeStack = v.scopeStack[:len(v.scopeStack)-1]
}

func (v *VariablesVisitor) currentScope() *sitter.Node {
	if len(v.scopeStack) == 0 {
		return &sitter.Node{}
	}
	return v.scopeStack[len(v.scopeStack)-1]
}

func (v *VariablesVisitor) addVariableDefinition(variableType VariableType, definitionNode, variableNameNode, variableValueNode *sitter.Node) {
	v.symbolTable.AddVariableDefinition(variableNameNode.Content(v.content), VariableDefinition{
		Value:        variableValueNode.Content(v.content),
		VariableType: variableType,
		Scope: sitter.Range{
			StartPoint: definitionNode.StartPoint(),
			EndPoint:   v.currentScope().EndPoint(),
			StartByte:  definitionNode.StartByte(),
			EndByte:    v.currentScope().EndByte(),
		},
		Range: sitter.Range{
			StartPoint: variableNameNode.StartPoint(),
			EndPoint:   variableValueNode.EndPoint(),
			StartByte:  variableNameNode.StartByte(),
			EndByte:    variableValueNode.EndByte(),
		},
	})
}

func (v *VariablesVisitor) addVariableUsage(node *sitter.Node) {
	v.symbolTable.AddVariableUsage(node.Content(v.content), sitter.Range{
		StartPoint: node.StartPoint(),
		EndPoint:   node.EndPoint(),
		StartByte:  node.StartByte(),
		EndByte:    node.EndByte(),
	})
}

func (v *VariablesVisitor) EnterContextShift(_ *sitter.Node, _ string) {}
func (v *VariablesVisitor) ExitContextShift(_ *sitter.Node)            {}
