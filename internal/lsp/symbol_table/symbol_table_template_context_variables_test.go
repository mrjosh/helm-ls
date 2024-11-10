package symboltable

import (
	"strings"
	"testing"

	templateast "github.com/mrjosh/helm-ls/internal/lsp/template_ast"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
)

func TestResolveVariablesInTemplateContext(t *testing.T) {
	tests := []struct {
		template    string
		templateCtx TemplateContext
		expectedCtx TemplateContext
		expectedErr error
	}{
		{"{{ $values := .Values }} {{ $values.te^st }}", TemplateContext{"$values", "test"}, TemplateContext{"Values", "test"}, nil},
		{"{{- range $type, $config := .Values.deployments }} {{ $config.te^st }}", TemplateContext{"$config", "test"}, TemplateContext{"Values", "deployments[]", "test"}, nil},
		{" {{ $values := .Values }} {{- range $type, $config := $values.deployments }} {{ $config.te^st }}", TemplateContext{"$config", "test"}, TemplateContext{"Values", "deployments[]", "test"}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			col := strings.Index(tt.template, "^")
			buf := strings.Replace(tt.template, "^", "", 1)
			ast := templateast.ParseAst(nil, []byte(tt.template))
			symbolTable := NewSymbolTable(ast, []byte(buf))

			result, err := symbolTable.ResolveVariablesInTemplateContext(
				tt.templateCtx, sitter.Range{StartByte: uint32(col), EndByte: uint32(col + 1)})

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCtx, result)
			}
		})
	}
}
