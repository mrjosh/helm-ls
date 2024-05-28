package lsp

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type TemplateContext []string

func (t TemplateContext) Format() string {
	return strings.Join(t, ".")
}

func (t TemplateContext) Copy() TemplateContext {
	return append(TemplateContext{}, t...)
}

func (t TemplateContext) Tail() TemplateContext {
	return t[1:]
}

func (t TemplateContext) IsVariable() bool {
	return len(t) > 0 && strings.HasPrefix(t[0], "$")
}

func (t TemplateContext) AppendSuffix(suffix string) TemplateContext {
	t[len(t)-1] = t[len(t)-1] + suffix
	return t
}

type SymbolTable struct {
	contexts            map[string][]sitter.Range
	contextsReversed    map[sitter.Range]TemplateContext
	includeDefinitions  map[string][]sitter.Range
	includeUseages      map[string][]sitter.Range
	variableDefinitions map[string][]VariableDefinition
}

func NewSymbolTable(ast *sitter.Tree, content []byte) *SymbolTable {
	s := &SymbolTable{
		contexts:            map[string][]sitter.Range{},
		contextsReversed:    map[sitter.Range]TemplateContext{},
		includeDefinitions:  map[string][]sitter.Range{},
		includeUseages:      map[string][]sitter.Range{},
		variableDefinitions: map[string][]VariableDefinition{},
	}
	s.parseTree(ast, content)
	return s
}

func (s *SymbolTable) AddTemplateContext(templateContext TemplateContext, pointRange sitter.Range) {
	if templateContext.IsVariable() && templateContext[0] == "$" {
		// $ is a special variable that resolves to the root context
		// we can just remove it from the template context
		templateContext = templateContext.Tail()
	}
	s.contexts[templateContext.Format()] = append(s.contexts[templateContext.Format()], pointRange)
	sliceCopy := make(TemplateContext, len(templateContext))
	copy(sliceCopy, templateContext)
	s.contextsReversed[pointRange] = sliceCopy
}

func (s *SymbolTable) GetTemplateContextRanges(templateContext TemplateContext) []sitter.Range {
	return s.contexts[templateContext.Format()]
}

func (s *SymbolTable) GetTemplateContext(pointRange sitter.Range) (TemplateContext, error) {
	result, ok := s.contextsReversed[pointRange]
	if !ok {
		return result, fmt.Errorf("no template context found")
	}
	// return a copy to never modify the original
	return result.Copy(), nil
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

func (s *SymbolTable) GetAllIncludeDefinitionsNames() (result []string) {
	for k := range s.includeDefinitions {
		result = append(result, k)
	}
	return result
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
			NewVariablesVisitor(s, content),
		},
	}

	v.visitNodesRecursiveWithScopeShift(rootNode)
}
