package lsp

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type TemplateContext []string

type SymbolTable struct {
	contexts           map[string][]sitter.Range
	contextsReversed   map[sitter.Range]TemplateContext
	includeDefinitions map[string][]sitter.Range
	includeUseages     map[string][]sitter.Range
}

func NewSymbolTable(ast *sitter.Tree, content []byte) *SymbolTable {
	s := &SymbolTable{
		contexts:           make(map[string][]sitter.Range),
		contextsReversed:   make(map[sitter.Range]TemplateContext),
		includeDefinitions: make(map[string][]sitter.Range),
		includeUseages:     make(map[string][]sitter.Range),
	}
	s.parseTree(ast, content)
	return s
}

func (s *SymbolTable) AddValue(templateContext []string, pointRange sitter.Range) {
	s.contexts[strings.Join(templateContext, ".")] = append(s.contexts[strings.Join(templateContext, ".")], pointRange)
	s.contextsReversed[pointRange] = templateContext
}

func (s *SymbolTable) GetValues(path []string) []sitter.Range {
	return s.contexts[strings.Join(path, ".")]
}

func (s *SymbolTable) GetTemplateContext(pointRange sitter.Range) TemplateContext {
	return s.contextsReversed[pointRange]
}

func (s *SymbolTable) AddIncludeDefinition(symbol string, pointRange sitter.Range) {
	s.includeDefinitions[symbol] = append(s.includeDefinitions[symbol], pointRange)
}

func (s *SymbolTable) AddIncludeReference(symbol string, pointRange sitter.Range) {
	s.includeUseages[symbol] = append(s.includeUseages[symbol], pointRange)
}

func (s *SymbolTable) GetIncludeDefinitions(symbol string) []sitter.Range {
	return s.includeDefinitions[symbol]
}

func (s *SymbolTable) GetIncludeReference(symbol string) []sitter.Range {
	result := s.includeUseages[symbol]
	definitions := s.includeDefinitions[symbol]
	return append(result, definitions...)
}

func (s *SymbolTable) parseTree(ast *sitter.Tree, content []byte) {
	rootNode := ast.RootNode()
	v := Visitors{
		symbolTable: s,
		visitors: []Visitor{
			NewValuesVisitor(s, content),
			NewIncludeDefinitionsVisitor(s, content),
		},
	}

	v.visitNodesRecursiveWithScopeShift(rootNode)
}
