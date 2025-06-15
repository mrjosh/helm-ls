package util

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

func TestPositionFinderVisitor(t *testing.T) {
	var node ast.Node

	yml := `
a: 
  World: something
b: c
`

	err := yaml.Unmarshal([]byte(yml), &node)
	assert.NoError(t, err)

	pv := &PositionFinderVisitor{
		position: protocol.Position{
			Line:      2,
			Character: 2,
		},
	}

	// Walk the AST with the position finder visitor.
	ast.Walk(pv, node)

	assert.True(t, pv.found)
	assert.NotNil(t, pv.result)

	// Check if the result is as expected.
	expectedResult := "World"
	assert.Equal(t, expectedResult, pv.result.String())
}
