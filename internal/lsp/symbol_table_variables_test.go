package lsp

import (
	"fmt"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
)

func TestGetVariableDefinition(t *testing.T) {
	testCases := []struct {
		template      string
		variableName  string
		accessRange   sitter.Range // the range must currently only contain the bytes
		expectedValue string
		expectedError error
	}{
		{
			template:     `{{ $x := "hello" }} {{ $x }}`,
			variableName: "$x",
			accessRange: sitter.Range{
				StartByte: 20,
				EndByte:   21,
			},
			expectedValue: `"hello"`,
			expectedError: nil,
		},
		{
			template: `
			{{ if true }}
			{{ $x := "hello" }} {{ $x }}
			{{ end }}
			{{ if false }}
			{{ $x := "goodby" }} {{ $x }}
			{{ end }}
			`,
			variableName: "$x",
			accessRange: sitter.Range{
				StartByte: 46,
				EndByte:   47,
			},
			expectedValue: `"hello"`,
			expectedError: nil,
		},
		{
			template: `
			{{ if true }}
			{{ $x := "hello" }} {{ $x }}
			{{ end }}
			{{ if false }}
			{{ $x := "goodby" }} {{ $x }}
			{{ end }}
			`,
			variableName: "$x",
			accessRange: sitter.Range{
				StartByte: 110,
				EndByte:   111,
			},
			expectedValue: `"goodby"`,
			expectedError: nil,
		},
		{
			template: `
			{{ if true }}
			{{ $x := "hello" }} {{ $x }}
			{{ end }}
			`,
			variableName: "$x",
			accessRange: sitter.Range{
				StartByte: 67,
				EndByte:   68,
			},
			expectedValue: ``,
			expectedError: fmt.Errorf("variable $x not found: variable not found"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.template, func(t *testing.T) {
			ast := ParseAst(nil, []byte(tC.template))
			symbolTable := NewSymbolTable(ast, []byte(tC.template))
			result, err := symbolTable.getVariableDefinition(tC.variableName, tC.accessRange)
			assert.Equal(t, tC.expectedError, err)
			assert.Equal(t, tC.expectedValue, result.Value)
		})
	}
}
