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
	ast := ParseAst(nil, []byte(template))

	logger.Println("RootNode:", ast.RootNode().String())

	node := FindRelevantChildNodeCompletion(ast.RootNode(), sitter.Point{
		Row:    0,
		Column: 11,
	})

	assert.Equal(t, sitter.Point{
		Row:    0,
		Column: 10,
	}, node.StartPoint())
	assert.Equal(t, sitter.Point{
		Row:    0,
		Column: 11,
	}, node.EndPoint())
}
