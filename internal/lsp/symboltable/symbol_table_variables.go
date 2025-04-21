package symboltable

import (
	"fmt"

	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
)

type VariableType int64

const (
	VariableTypeAssigment VariableType = iota
	VariableTypeRangeKeyOrIndex
	VariableTypeRangeValue
)

type VariableDefinition struct {
	Value        string
	VariableType VariableType
	Scope        sitter.Range
	Range        sitter.Range
}

func (s *SymbolTable) AddVariableDefinition(symbol string, variableDefinition VariableDefinition) {
	s.variableDefinitions[symbol] = append(s.variableDefinitions[symbol], variableDefinition)
}

func (s *SymbolTable) AddVariableUsage(symbol string, pointRange sitter.Range) {
	s.variableUsages[symbol] = append(s.variableUsages[symbol], pointRange)
}

func (s *SymbolTable) getVariableDefinition(name string, accessRange sitter.Range) (VariableDefinition, error) {
	definitions, ok := s.variableDefinitions[name]
	if !ok || len(definitions) == 0 {
		return VariableDefinition{}, fmt.Errorf("variable %s not found", name)
	}

	definition, err := findDefinitionForRange(definitions, accessRange)
	if err != nil {
		return VariableDefinition{}, fmt.Errorf("variable %s not found: %s", name, err.Error())
	}

	return definition, nil
}

func (s *SymbolTable) GetAllVariableDefinitions() (result map[string][]VariableDefinition) {
	result = map[string][]VariableDefinition{}
	for name, definitions := range s.variableDefinitions {
		result[name] = append(
			[]VariableDefinition{},
			definitions...)
	}
	return result
}

func (s *SymbolTable) GetVariableDefinitionForNode(node *sitter.Node, content []byte) (VariableDefinition, error) {
	name, err := getVariableName(node, content)
	if err != nil {
		return VariableDefinition{}, err
	}
	return s.getVariableDefinition(name, node.Range())
}

func (s *SymbolTable) GetVariableReferencesForNode(node *sitter.Node, content []byte) (ranges []sitter.Range, err error) {
	name, err := getVariableName(node, content)
	if err != nil {
		return []sitter.Range{}, err
	}
	definition, err := s.GetVariableDefinitionForNode(node, content)
	if err != nil {
		return []sitter.Range{}, err
	}
	usages := s.variableUsages[name]

	for _, usage := range usages {
		if util.RangeContainsRange(definition.Scope, usage) {
			ranges = append(ranges, usage)
		}
	}
	return append(ranges, definition.Range), nil
}

func getVariableName(node *sitter.Node, content []byte) (string, error) {
	if node == nil {
		return "", fmt.Errorf("Cannot get variable definition for node")
	}
	if node.Type() == gotemplate.NodeTypeIdentifier {
		node = node.Parent()
	}
	if node.Type() != gotemplate.NodeTypeVariable {
		return "", fmt.Errorf("Node is not a variable but is of type %s", node.Type())
	}
	return node.Content(content), nil
}

func findDefinitionForRange(definitions []VariableDefinition, accessRange sitter.Range) (VariableDefinition, error) {
	for _, definition := range definitions {
		if util.RangeContainsRange(definition.Scope, accessRange) {
			return definition, nil
		}
	}
	return VariableDefinition{}, fmt.Errorf("variable not found")
}
