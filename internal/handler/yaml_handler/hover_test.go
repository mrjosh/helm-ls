package yamlhandler

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/mrjosh/helm-ls/internal/testutil"
	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestHover(t *testing.T) {
	testCases := []struct {
		desc       string
		filepath   string
		markedLine string

		expected      string
		expectedError string
	}{
		{
			"Hover on mapping node",
			"../../../testdata/example/values.yaml",
			"replica^Count: 1",
			"\n\nreplicaCount\n\n",
			"",
		},
		{
			"Hover on nested mapping node",
			"../../../testdata/example/values.yaml",
			"pathTy^pe: ImplementationSpecific",
			"\n\ningress.hosts[].paths[].pathType\n\n",
			"",
		},
		{
			"Hover on nested mapping node, beginning of node",
			"../../../testdata/example/values.yaml",
			"^pathType: ImplementationSpecific",
			"\n\ningress.hosts[].paths[].pathType\n\n",
			"",
		},
		{
			"Hover on nested mapping node, end of node",
			"../../../testdata/example/values.yaml",
			"pathTyp^e: ImplementationSpecific",
			"\n\ningress.hosts[].paths[].pathType\n\n",
			"",
		},
		{
			"Hover on nested mapping node, at colon",
			"../../../testdata/example/values.yaml",
			"pathType:^ ImplementationSpecific",
			"",
			"YAML node not found for position {55 19}",
		},
		{
			"Hover on nested value node",
			"../../../testdata/example/values.yaml",
			"pathType: Imple^mentationSpecific",
			"\n\ningress.hosts[].paths[].pathType\n\n",
			"",
		},
		{
			"Hover on value defined in multiple files",
			"../../../testdata/dependenciesExample/values.yaml",
			`"wi^th.dot": hi`,
			fmt.Sprintf("\n\nglobal.with.dot\n\n### %s\n```yaml\nhi\n```\n", filepath.Join("..", "..", "..", "testdata", "dependenciesExample", "values.a.yaml")),
			"",
		},
		{
			"Hover on global",
			"../../../testdata/dependenciesExample/values.yaml",
			`glob^al:`,
			fmt.Sprintf("\n\nglobal\n\n### %s\n```yaml\nglobalFromFileA: hi\nwith.dot: hi\n```\n### %s\n```yaml\nglobalFromSubchart: works\nsubchart: works\n```\n",
				filepath.Join("..", "..", "..", "testdata", "dependenciesExample", "values.a.yaml"), filepath.Join("..", "..", "..", "testdata", "dependenciesExample", "charts", "subchartexample", "values.yaml")),
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			h, fileContent := setupYamlHandlerTest(t, tc.filepath, false)
			pos, found := testutil.GetPositionOfMarkedLineInFile(fileContent, tc.markedLine, "^")
			assert.True(t, found)

			result, err := h.Hover(context.Background(), &lsp.HoverParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{
						URI: uri.File(tc.filepath),
					},
					Position: pos,
				},
			})

			assert.NotNil(t, result)
			assert.Equal(t, tc.expected, result.Contents.Value)
			if tc.expectedError == "" {
				assert.Nil(t, err)
				if err != nil {
					t.Errorf("Expected no error, got %v", err.Error())
				}
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
