package lsp

import (
	"strings"

	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
)

type SymbolTable struct {
	values             map[string][]sitter.Range
	includeDefinitions map[string]sitter.Range
}

func NewSymbolTable(ast *sitter.Tree) *SymbolTable {
	return &SymbolTable{
		values:             make(map[string][]sitter.Range),
		includeDefinitions: make(map[string]sitter.Range),
	}
}

func (s *SymbolTable) AddValue(symbol string, pointRange sitter.Range) {
	s.values[symbol] = append(s.values[symbol], pointRange)
}

func (s *SymbolTable) AddIncludeDefinition(symbol string, pointRange sitter.Range) {
	s.includeDefinitions[symbol] = pointRange
}

func (s *SymbolTable) parseTree(ast *sitter.Tree, content []byte) {
	rootNode := ast.RootNode()

	v := Visitors{
		symbolTable: s,
		visitors: []Visitor{
			&ValuesVisitor{
				currentContext: []string{},
				stashedContext: [][]string{},
				symbolTable:    s,
				content:        content,
			},
			&IncludeDefinitionsVisitor{
				symbolTable: s,
				content:     content,
			},
		},
	}

	v.visitNodesRecursiveWithScopeShift(rootNode)
}

type IncludeDefinitionsVisitor struct {
	symbolTable *SymbolTable
	content     []byte
}

func (v *IncludeDefinitionsVisitor) Enter(node *sitter.Node) {
	if node.Type() != gotemplate.NodeTypeDefineAction {
		return
	}
	v.symbolTable.AddIncludeDefinition(node.ChildByFieldName("name").Content(v.content), getRangeForNode(node))
}

func (v *IncludeDefinitionsVisitor) Exit(node *sitter.Node)                           {}
func (v *IncludeDefinitionsVisitor) EnterScopeShift(node *sitter.Node, suffix string) {}
func (v *IncludeDefinitionsVisitor) ExitScopeShift(node *sitter.Node)                 {}

type ValuesVisitor struct {
	currentContext []string
	stashedContext [][]string
	symbolTable    *SymbolTable
	content        []byte
}

func (v *ValuesVisitor) Enter(node *sitter.Node) {
	switch node.Type() {
	case gotemplate.NodeTypeDot:
		v.symbolTable.AddValue(strings.Join(v.currentContext, "."), getRangeForNode(node))
	case gotemplate.NodeTypeFieldIdentifier:
		content := node.Content(v.content)
		v.symbolTable.AddValue(strings.Join(append(v.currentContext, content), "."), getRangeForNode(node))
	case gotemplate.NodeTypeField:
		content := node.ChildByFieldName("name").Content(v.content)
		value := strings.Join(append(v.currentContext, content), ".")
		v.symbolTable.AddValue(value, getRangeForNode(node))
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

func (v *ValuesVisitor) EnterScopeShift(node *sitter.Node, suffix string) {
	switch node.Type() {
	case gotemplate.NodeTypeFieldIdentifier:
		content := node.Content(v.content) + suffix
		v.currentContext = append(v.currentContext, content)
	case gotemplate.NodeTypeField:
		content := node.ChildByFieldName("name").Content(v.content) + suffix
		v.currentContext = append(v.currentContext, content)
	case gotemplate.NodeTypeSelectorExpression:
		s := getScopeForSelectorExpression(node, v.content)
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

func (v *ValuesVisitor) ExitScopeShift(node *sitter.Node) {
	switch node.Type() {
	case gotemplate.NodeTypeField, gotemplate.NodeTypeFieldIdentifier:
		v.currentContext = v.currentContext[:len(v.currentContext)-1]
	case gotemplate.NodeTypeSelectorExpression:
		s := getScopeForSelectorExpression(node, v.content)
		if len(s) > 0 && s[0] == "$" {
			v.currentContext = v.stashedContext[len(v.stashedContext)-1]
			v.stashedContext = v.stashedContext[:len(v.stashedContext)-1]
			s = s[1:]
		} else {
			v.currentContext = v.currentContext[:len(v.currentContext)-len(s)]
		}
	}
}

func getScopeForSelectorExpression(node *sitter.Node, content []byte) []string {
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
	operandScope := getScopeForSelectorExpression(operand, content)
	field := node.ChildByFieldName("field")
	if field == nil {
		return operandScope
	}
	fieldScope := field.Content(content)

	return append(operandScope, fieldScope)
}
