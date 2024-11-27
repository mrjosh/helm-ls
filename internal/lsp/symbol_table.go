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

// Return everything except the first context
func (t TemplateContext) Tail() TemplateContext {
	return t[1:]
}

func (t TemplateContext) IsVariable() bool {
	return len(t) > 0 && strings.HasPrefix(t[0], "$")
}

// Adds a suffix to the last context
func (t TemplateContext) AppendSuffix(suffix string) TemplateContext {
	t[len(t)-1] = t[len(t)-1] + suffix
	return t
}

func NewTemplateContext(string string) TemplateContext {
	if string == "." {
		return TemplateContext{}
	}
	splitted := strings.Split(string, ".")
	if len(splitted) > 0 && splitted[0] == "" {
		return splitted[1:]
	}
	return splitted
}

type SymbolTable struct {
	contexts            map[string][]sitter.Range
	contextsReversed    map[sitter.Range]TemplateContext
	includeDefinitions  map[string][]sitter.Range
	includeUsages       map[string][]sitter.Range
	variableDefinitions map[string][]VariableDefinition
	variableUsages      map[string][]sitter.Range
}

func NewSymbolTable(ast *sitter.Tree, content []byte) *SymbolTable {
	s := &SymbolTable{
		contexts:            map[string][]sitter.Range{},
		contextsReversed:    map[sitter.Range]TemplateContext{},
		includeDefinitions:  map[string][]sitter.Range{},
		includeUsages:       map[string][]sitter.Range{},
		variableDefinitions: map[string][]VariableDefinition{},
		variableUsages:      map[string][]sitter.Range{},
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
		return result, fmt.Errorf("No template context found for range %v", pointRange)
	}
	// return a copy to never modify the original
	return s.ResolveVariablesInTemplateContext(result.Copy(), pointRange)
}

func (s *SymbolTable) AddIncludeDefinition(symbol string, pointRange sitter.Range) {
	s.includeDefinitions[symbol] = append(s.includeDefinitions[symbol], pointRange)
}

func (s *SymbolTable) AddIncludeReference(symbol string, pointRange sitter.Range) {
	s.includeUsages[symbol] = append(s.includeUsages[symbol], pointRange)
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
	result := s.includeUsages[symbol]
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
