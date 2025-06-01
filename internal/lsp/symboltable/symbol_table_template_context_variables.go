package symboltable

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

func (s *SymbolTable) ResolveVariablesInTemplateContext(templateContext TemplateContext, pointRange sitter.Range) (TemplateContext, error) {
	if !templateContext.IsVariable() {
		return templateContext, nil
	}

	variableName := templateContext[0]
	if variableName == "$" {
		return templateContext.Tail(), nil
	}

	variableDefinitions := s.variableDefinitions[variableName]

	if len(variableDefinitions) == 0 {
		return templateContext, fmt.Errorf("variable %s not found", variableName)
	}

	definition, err := findDefinitionForRange(variableDefinitions, pointRange)
	if err != nil {
		return templateContext, fmt.Errorf("variable %s not found %e", variableName, err)
	}

	prefix := getPrefixTemplateContextForVariable(definition)

	return s.ResolveVariablesInTemplateContext(append(prefix, templateContext.Tail()...), pointRange)
}

func getPrefixTemplateContextForVariable(definition VariableDefinition) TemplateContext {
	prefix := NewTemplateContext(definition.Value)
	if definition.VariableType == VariableTypeRangeValue && len(prefix) > 0 {
		prefix[len(prefix)-1] = prefix[len(prefix)-1] + "[]"
	}
	return prefix
}
