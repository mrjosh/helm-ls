package util

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

// func TestPositionFinderVisitor(t *testing.T) {
// 	var node ast.Node
//
// 	yml := `
// a:
//   World: something
// b: c
// `
//
// 	err := yaml.Unmarshal([]byte(yml), &node)
// 	assert.NoError(t, err)
//
// 	pv := &PositionFinderVisitor{
// 		position: protocol.Position{
// 			Line:      2,
// 			Character: 2,
// 		},
// 	}
//
// 	ast.Walk(pv, node)
//
// 	assert.True(t, pv.found)
// 	assert.NotNil(t, pv.result)
//
// 	expectedResult := "World"
// 	assert.Equal(t, expectedResult, pv.result.String())
// }

func TestPositionFinderVisitor(t *testing.T) {
	const yamlDoc = `
root:
  parent1:
    childA: valueA
    childB: valueB
  parent2:
    - listItem1
    - listItem2
  scalarKey: scalarValue
`

	var node ast.Node
	err := yaml.Unmarshal([]byte(yamlDoc), &node)
	assert.NoError(t, err)

	tests := []struct {
		name       string
		line       uint32
		character  uint32
		wantFound  bool
		wantResult string
	}{
		{
			name:       "inside parent1 key",
			line:       2, // line numbers are 0-based if your visitor uses 0-based
			character:  4,
			wantFound:  true,
			wantResult: "parent1",
		},
		{
			name:       "inside childA key",
			line:       3,
			character:  6,
			wantFound:  true,
			wantResult: "childA",
		},
		{
			name:      "inside list",
			line:      6,
			character: 4,
			wantFound: false,
		},
		{
			name:       "inside list item 1",
			line:       6,
			character:  12,
			wantFound:  true,
			wantResult: "listItem1",
		},
		{
			name:       "on scalar value",
			line:       8,
			character:  14,
			wantFound:  true,
			wantResult: "scalarValue",
		},
		{
			name:       "outside any node",
			line:       0,
			character:  0,
			wantFound:  false,
			wantResult: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fresh visitor for each test
			pv := &PositionFinderVisitor{
				position: protocol.Position{
					Line:      tc.line,
					Character: tc.character,
				},
			}

			ast.Walk(pv, node)

			assert.Equal(t, tc.wantFound, pv.found, "found flag")

			if tc.wantFound {
				assert.NotNil(t, pv.result, "result should not be nil")
				assert.Equal(t, tc.wantResult, pv.result.String(), "node value")
			} else {
				assert.Nil(t, pv.result, "result should be nil when not found")
			}
		})
	}
}
