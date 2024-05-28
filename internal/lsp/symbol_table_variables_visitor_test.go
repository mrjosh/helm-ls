package lsp

import (
	"fmt"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
)

func TestSymbolTableForVariableDefinitions(t *testing.T) {
	type testCase struct {
		template                    string
		expectedVariableDefinitions map[string][]VariableDefinition
	}

	testCases := []testCase{
		{
			template: `{{- $root := . -}} `,
			expectedVariableDefinitions: map[string][]VariableDefinition{
				"$root": {
					{
						Value:        ".",
						VariableType: VariableTypeAssigment,
						Scope: sitter.Range{
							StartPoint: sitter.Point{
								Row:    0,
								Column: 14,
							},
							EndPoint: sitter.Point{
								Row:    0,
								Column: 19,
							},
							StartByte: 14,
							EndByte:   19,
						},
					},
				},
			},
		},
		{
			template: `
		      {{- range $type, $config := $root.Values.deployments }}
						{{- .InLoop }}
					{{- end }}
			`,
			expectedVariableDefinitions: map[string][]VariableDefinition{
				"$type": {
					{
						Value:        "$root.Values.deployments",
						VariableType: VariableTypeRangeKeyOrIndex,
						Scope: sitter.Range{
							StartPoint: sitter.Point{
								Row:    1,
								Column: 60,
							},
							EndPoint: sitter.Point{
								Row:    3,
								Column: 15,
							},
							StartByte: 61,
							EndByte:   101,
						},
					},
				},
				"$config": {
					{
						Value:        "$root.Values.deployments",
						VariableType: VariableTypeRangeValue,
						Scope: sitter.Range{
							StartPoint: sitter.Point{
								Row:    1,
								Column: 60,
							},
							EndPoint: sitter.Point{
								Row:    3,
								Column: 15,
							},
							StartByte: 61,
							EndByte:   101,
						},
					},
				},
			},
		},
		{
			template: `{{ $x := .Values }}{{ $x.test }}{{ .Values.test }}`,
			expectedVariableDefinitions: map[string][]VariableDefinition{
				"$x": {
					{
						Value:        ".Values",
						VariableType: VariableTypeAssigment,
						Scope: sitter.Range{
							StartPoint: sitter.Point{
								Row:    0,
								Column: 16,
							},
							EndPoint: sitter.Point{
								Row:    0,
								Column: 50,
							},
							StartByte: 16,
							EndByte:   50,
						},
					},
				},
			},
		},
	}

	for _, v := range testCases {
		t.Run(v.template, func(t *testing.T) {
			ast := ParseAst(nil, v.template)
			symbolTable := NewSymbolTable(ast, []byte(v.template))
			assert.Equal(t, v.expectedVariableDefinitions, symbolTable.variableDefinitions,
				fmt.Sprintf("Ast was %s", ast.RootNode()))
		})
	}
}
