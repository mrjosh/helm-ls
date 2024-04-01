package lsp

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type SymbolTable struct {
	values             map[string][]sitter.Range
	includeDefinitions map[string][]sitter.Range
	includeReferences  map[string][]sitter.Range
}

func NewSymbolTable(ast *sitter.Tree, content []byte) *SymbolTable {
	s := &SymbolTable{
		values:             make(map[string][]sitter.Range),
		includeDefinitions: make(map[string][]sitter.Range),
		includeReferences:  make(map[string][]sitter.Range),
	}
	s.parseTree(ast, content)
	return s
}

func (s *SymbolTable) AddValue(path []string, pointRange sitter.Range) {
	s.values[strings.Join(path, ".")] = append(s.values[strings.Join(path, ".")], pointRange)
}

func (s *SymbolTable) GetValues(path []string) []sitter.Range {
	return s.values[strings.Join(path, ".")]
}

func (s *SymbolTable) AddIncludeDefinition(symbol string, pointRange sitter.Range) {
	s.includeDefinitions[symbol] = append(s.includeDefinitions[symbol], pointRange)
}

func (s *SymbolTable) AddIncludeReference(symbol string, pointRange sitter.Range) {
	s.includeReferences[symbol] = append(s.includeReferences[symbol], pointRange)
}

func (s *SymbolTable) GetIncludeDefinitions(symbol string) ([]sitter.Range, bool) {
	result, ok := s.includeDefinitions[symbol]
	if !ok {
		return []sitter.Range{}, false
	}
	return result, true
}

func (s *SymbolTable) GetIncludeReference(symbol string) ([]sitter.Range, bool) {
	result, ok := s.includeReferences[symbol]
	if !ok {
		return []sitter.Range{}, false
	}
	return result, true
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
