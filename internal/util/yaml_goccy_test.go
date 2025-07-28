package util

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/mrjosh/helm-ls/internal/testutil"
	"github.com/stretchr/testify/assert"
)

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
  inlineList: [1, 2, 3]
`

	var node ast.Node
	err := yaml.Unmarshal([]byte(yamlDoc), &node)
	assert.NoError(t, err)

	tests := []struct {
		name       string
		markedLine string
		wantFound  bool
		wantResult string
	}{
		{
			name:       "inside parent1 key",
			markedLine: "paren^t1",
			wantFound:  true,
			wantResult: "parent1",
		},
		{
			name:       "inside childA key",
			markedLine: "ch^ildA:",
			wantFound:  true,
			wantResult: "childA",
		},
		{
			name:       "around childA key left",
			markedLine: "^childA:",
			wantFound:  true,
			wantResult: "childA",
		},
		{
			name:       "around childA key right",
			markedLine: "child^A:",
			wantFound:  true,
			wantResult: "childA",
		},
		{
			name:       "inside list",
			markedLine: "^- listItem2",
			wantFound:  false,
		},
		{
			name:       "inside list item 1",
			markedLine: "- listI^tem1",
			wantFound:  true,
			wantResult: "listItem1",
		},
		{
			name:       "on scalar key left",
			markedLine: "^scalarKey: scalarValue",
			wantFound:  true,
			wantResult: "scalarKey",
		},
		{
			name:       "on scalar key left right",
			markedLine: "scalarKe^y: scalarValue",
			wantFound:  true,
			wantResult: "scalarKey",
		},
		{
			name:       "on scalar key left right",
			markedLine: "scalarKey^: scalarValue",
			wantFound:  false,
			wantResult: "",
		},
		{
			name:       "on scalar value",
			markedLine: "scalarKey: scala^rValue",
			wantFound:  true,
			wantResult: "scalarValue",
		},
		{
			name:       "on scalar value left border",
			markedLine: "scalarKey: ^scalarValue",
			wantFound:  true,
			wantResult: "scalarValue",
		},
		{
			name:       "on scalar value right border",
			markedLine: "scalarKey: scalarValue^",
			wantFound:  true,
			wantResult: "scalarValue",
		},
		{
			name:       "outside any node",
			wantFound:  false,
			wantResult: "",
		},
		{
			name:       "on inline list",
			markedLine: "inlineList: [^1, 2, 3]",
			wantFound:  true,
			wantResult: "1",
		},
		{
			name:       "on inline list comma",
			markedLine: "inlineList: [1^, 2, 3]",
			wantFound:  false,
			wantResult: "",
		},
		{
			name:       "on inline list end",
			markedLine: "inlineList: [1, 2, 3^]",
			wantFound:  false,
			wantResult: "",
		},
		{
			name:       "on inline list start",
			markedLine: "inlineList: ^[1, 2, 3]",
			wantFound:  false,
			wantResult: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pos, found := testutil.GetPositionOfMarkedLineInFile(yamlDoc, tc.markedLine, "^")
			assert.True(t, found)
			pv := &PositionFinderVisitor{
				position: pos,
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
