package lsp

import (
	"fmt"

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

func (s *SymbolTable) getVariableDefinition(name string, accessRange sitter.Range) (VariableDefinition, error) {
	definitions, ok := s.variableDefinitions[name]
	if !ok || len(definitions) == 0 {
		return VariableDefinition{}, fmt.Errorf("variable %s not found", name)
	}
	for _, definition := range definitions {
		if util.RangeContainsRange(definition.Scope, accessRange) {
			return definition, nil
		}
	}
	return VariableDefinition{}, fmt.Errorf("variable %s not found", name)
}

func (s *SymbolTable) GetVariableDefinitionForNode(node *sitter.Node) (VariableDefinition, error) {
	if node == nil {
		return VariableDefinition{}, fmt.Errorf("Cannot get variable definition for node")
	}
	return VariableDefinition{}, fmt.Errorf("Cannot get variable definition for node")
}
