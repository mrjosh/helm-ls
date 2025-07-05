package yamlhandler

import (
	"context"
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
			"$.replicaCount",
			"",
		},
		{
			"Hover on nested mapping node",
			"../../../testdata/example/values.yaml",
			"pathTy^pe: ImplementationSpecific",
			"$.ingress.hosts[0].paths[0].pathType",
			"",
		},
		{
			"Hover on nested mapping node, beginning of node",
			"../../../testdata/example/values.yaml",
			"^pathType: ImplementationSpecific",
			"$.ingress.hosts[0].paths[0].pathType",
			"",
		},
		{
			"Hover on nested mapping node, end of node",
			"../../../testdata/example/values.yaml",
			"pathType^: ImplementationSpecific",
			"$.ingress.hosts[0].paths[0].pathType",
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
			"$.ingress.hosts[0].paths[0].pathType",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			h, fileContent := setupYamlHandlerTest(t, tc.filepath)
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
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
