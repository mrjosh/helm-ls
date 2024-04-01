package lsp

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
)

func TestSymbolTableForIncludeDefinitions(t *testing.T) {
	content := `
	{{ define "foo" }}
	{{ .Values.global. }}
	{{ end }}

	{{ define "bar" }}
	{{ .Values.global. }}	
	{{ end }}
	`

	ast := ParseAst(nil, content)

	symbolTable := NewSymbolTable(ast)

	symbolTable.parseTree(ast, []byte(content))

	assert.Len(t, symbolTable.includeDefinitions, 2)

	// TODO: remove the double quotes
	assert.Equal(t, symbolTable.includeDefinitions["\"bar\""], sitter.Range{
		StartPoint: sitter.Point{
			Row:    5,
			Column: 0,
		},
		EndPoint: sitter.Point{
			Row:    7,
			Column: 10,
		},
		StartByte: 56,
		EndByte:   110,
	})
}

func TestSymbolTableForValues(t *testing.T) {
	content := `
{{ with .Values.with.something }}
{{ .test2 }}
{{ end }}

{{ .Test }}
{{ .Values.with.something }}


{{ range .list }}
	{{ . }}
	{{ .listinner }}
	{{ $.dollar }}
	{{ range .nested }}
		{{ .nestedinList }}
	{{ end }}
	{{ range $.Values.dollar }}
		{{ .nestedinList }}
	{{ end }}
{{ end }}

{{ .Test }}
`

	ast := ParseAst(nil, content)

	symbolTable := NewSymbolTable(ast)

	symbolTable.parseTree(ast, []byte(content))

	for k, v := range symbolTable.values {
		logger.Println(k, v)
	}

	assert.False(t, true)
}
