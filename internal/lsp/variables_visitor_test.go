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
						Scope:        sitter.Range{StartPoint: sitter.Point{Row: 0, Column: 4}, EndPoint: sitter.Point{Row: 0, Column: 19}, StartByte: 4, EndByte: 19},
						Range:        sitter.Range{StartPoint: sitter.Point{Row: 0, Column: 4}, EndPoint: sitter.Point{Row: 0, Column: 14}, StartByte: 4, EndByte: 14},
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
						Scope:        sitter.Range{StartPoint: sitter.Point{Row: 1, Column: 18}, EndPoint: sitter.Point{Row: 3, Column: 15}, StartByte: 19, EndByte: 101},
						Range:        sitter.Range{StartPoint: sitter.Point{Row: 1, Column: 18}, EndPoint: sitter.Point{Row: 1, Column: 60}, StartByte: 19, EndByte: 61},
					},
				},
				"$config": {
					{
						Value:        "$root.Values.deployments",
						VariableType: VariableTypeRangeValue,
						Scope:        sitter.Range{StartPoint: sitter.Point{Row: 1, Column: 18}, EndPoint: sitter.Point{Row: 3, Column: 15}, StartByte: 19, EndByte: 101},
						Range:        sitter.Range{StartPoint: sitter.Point{Row: 1, Column: 25}, EndPoint: sitter.Point{Row: 1, Column: 60}, StartByte: 26, EndByte: 61},
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
						Scope:        sitter.Range{StartPoint: sitter.Point{Row: 0, Column: 3}, EndPoint: sitter.Point{Row: 0, Column: 50}, StartByte: 3, EndByte: 50},
						Range:        sitter.Range{StartPoint: sitter.Point{Row: 0, Column: 3}, EndPoint: sitter.Point{Row: 0, Column: 16}, StartByte: 3, EndByte: 16},
					},
				},
			},
		},
	}

	for _, v := range testCases {
		t.Run(v.template, func(t *testing.T) {
			ast := ParseAst(nil, []byte(v.template))
			symbolTable := NewSymbolTable(ast, []byte(v.template))
			assert.Equal(t, v.expectedVariableDefinitions, symbolTable.variableDefinitions,
				fmt.Sprintf("Ast was %s", ast.RootNode()))
		})
	}
}

func TestSymbolTableForVariableUsages(t *testing.T) {
	type testCase struct {
		template               string
		expectedVariableUsages map[string][]sitter.Range
	}

	testCases := []testCase{
		{
			template: `{{ $x := .Values }}{{ $x.test }}{{ .Values.test }}`,
			expectedVariableUsages: map[string][]sitter.Range{
				"$x": {
					sitter.Range{StartPoint: sitter.Point{Row: 0, Column: 22}, EndPoint: sitter.Point{Row: 0, Column: 24}, StartByte: 22, EndByte: 24},
				},
			},
		},
	}

	for _, v := range testCases {
		t.Run(v.template, func(t *testing.T) {
			ast := ParseAst(nil, []byte(v.template))
			symbolTable := NewSymbolTable(ast, []byte(v.template))
			assert.Equal(t, v.expectedVariableUsages, symbolTable.variableUsages,
				fmt.Sprintf("Ast was %s", ast.RootNode()))
		})
	}
}
