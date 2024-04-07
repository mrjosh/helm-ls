package lsp

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type TemplateContext []string

func (t TemplateContext) Format() string {
	return strings.Join(t, ".")
}

func (t TemplateContext) Tail() TemplateContext {
	return t[1:]
}

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

func (s *SymbolTable) AddTemplateContext(templateContext TemplateContext, pointRange sitter.Range) {
	s.contexts[templateContext.Format()] = append(s.contexts[strings.Join(templateContext, ".")], pointRange)
	sliceCopy := make(TemplateContext, len(templateContext))
	copy(sliceCopy, templateContext)
	s.contextsReversed[pointRange] = sliceCopy
}

func (s *SymbolTable) GetTemplateContextRanges(templateContext TemplateContext) []sitter.Range {
	return s.contexts[templateContext.Format()]
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
			NewTemplateContextVisitor(s, content),
			NewIncludeDefinitionsVisitor(s, content),
		},
	}

	v.visitNodesRecursiveWithScopeShift(rootNode)
}
