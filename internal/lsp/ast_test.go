package lsp

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
)

func TestFindRelevantChildNodeCompletio(t *testing.T) {

	template := `{{ .Values. }}
{{ .Values.re }}

{{ toY }}

{{ .Chart.N }}
{{ . }}
`
	ast := ParseAst(nil, template)

	logger.Println("RootNode:", ast.RootNode().String())

	node := FindRelevantChildNodeCompletion(ast.RootNode(), sitter.Point{
		Row:    0,
		Column: 11,
	})

	assert.Equal(t, node.StartPoint(), sitter.Point{
		Row:    0,
		Column: 11,
	})
	assert.Equal(t, node.EndPoint(), sitter.Point{
		Row:    0,
		Column: 11,
	})
}
