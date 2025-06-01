package lsp

import (
	"strings"
	"testing"

	templateast "github.com/mrjosh/helm-ls/internal/lsp/template_ast"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

func TestFindRelevantChildNodeCompletion(t *testing.T) {
	tests := []struct {
		template    string
		startColumn uint32
		endColumn   uint32
		nodeType    string
		nodeContent string
	}{
		{
			template:    `{{ .Values.^test }}`,
			startColumn: 10,
			endColumn:   11,
			nodeType:    ".",
			nodeContent: ".",
		},
		{
			template:    `{{ .Values.^ }}`,
			startColumn: 10,
			endColumn:   11,
			nodeType:    ".",
			nodeContent: ".",
		},
		{
			template:    "{{ .Bad.^ }}",
			startColumn: 7,
			endColumn:   8,
			nodeType:    ".",
			nodeContent: ".",
		},
		{
			template:    "this is some additional text {{ .Bad.^ }}",
			startColumn: 36,
			endColumn:   37,
			nodeType:    ".",
			nodeContent: ".",
		},
		{
			template:    `{{ .Values.te^st }}`,
			startColumn: 11,
			endColumn:   15,
			nodeType:    "field_identifier",
			nodeContent: "test",
		},
		{
			template:    `{{ .Values.t^est }}`,
			startColumn: 11,
			endColumn:   15,
			nodeType:    "field_identifier",
			nodeContent: "test",
		},
		{
			template:    `{{ print^ }}`,
			startColumn: 3,
			endColumn:   8,
			nodeType:    "identifier",
			nodeContent: "print",
		},
		{
			template:    `{{ pri^nt }}`,
			startColumn: 3,
			endColumn:   8,
			nodeType:    "identifier",
			nodeContent: "print",
		},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			position, content := getPositionForMarkedTestLine(t, tt.template)

			ast := templateast.ParseAst(nil, []byte(content))
			t.Logf("RootNode: %s", ast.RootNode().String())
			node := templateast.NestedNodeAtPositionForCompletion(ast, position)

			assert.Equal(t, tt.nodeContent, node.Content([]byte(content)))
			assert.Equal(t, tt.nodeType, node.Type())

			assert.Equal(t, sitter.Point{Column: tt.startColumn}, node.StartPoint())
			assert.Equal(t, sitter.Point{Column: tt.endColumn}, node.EndPoint())

			t.Log(node.Content([]byte(content)))
		})
	}
}

func getPositionForMarkedTestLine(t *testing.T, buf string) (protocol.Position, string) {
	t.Helper()
	col := strings.Index(buf, "^")
	if col == -1 {
		t.Fatalf("Missing '^' in %s", buf)
	}

	buf = strings.Replace(buf, "^", "", 1)
	pos := protocol.Position{Line: 0, Character: uint32(col)}
	return pos, buf
}
