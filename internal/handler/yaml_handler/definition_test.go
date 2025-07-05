package yamlhandler

import (
	"context"
	"testing"

	"github.com/mrjosh/helm-ls/internal/testutil"
	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestDefintion(t *testing.T) {
	testCases := []struct {
		desc       string
		filepath   string
		markedLine string

		expected      []testutil.ExpectedDefinitionResult
		expectedError string
	}{
		{
			"Only defined in same file",
			"../../../testdata/example/values.yaml",
			"replica^Count: 1",
			[]testutil.ExpectedDefinitionResult{},
			"",
		},
		{
			"Defined in multiple files in same chart",
			"../../../testdata/dependenciesExample/values.a.yaml",
			"ima^ge:",
			[]testutil.ExpectedDefinitionResult{
				{
					Filepath:   "../../../testdata/dependenciesExample/values.b.yaml",
					MarkedLine: "§image§:",
				},
				{
					Filepath:   "../../../testdata/dependenciesExample/values.yaml",
					MarkedLine: "§image§:",
				},
			},
			"",
		},
		{
			"From parent to subchart",
			"../../../testdata/dependenciesExample/values.yaml",
			"subchartWithout^Global: worksToo",

			[]testutil.ExpectedDefinitionResult{
				{
					Filepath:   "../../../testdata/dependenciesExample/charts/subchartexample/values.yaml",
					MarkedLine: "§subchartWithoutGlobal§: works",
				},
			},
			"",
		},
		{
			"From parent to subchart global",
			"../../../testdata/dependenciesExample/values.yaml",
			"  subch^art: works",

			[]testutil.ExpectedDefinitionResult{
				{
					Filepath:   "../../../testdata/dependenciesExample/charts/subchartexample/values.yaml",
					MarkedLine: "  §subchart§: works",
				},
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			h, fileContent := setupYamlHandlerTest(t, tc.filepath)
			pos, found := testutil.GetPositionOfMarkedLineInFile(fileContent, tc.markedLine, "^")
			assert.True(t, found)

			result, err := h.Definition(context.Background(), &lsp.DefinitionParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{
						URI: uri.File(tc.filepath),
					},
					Position: pos,
				},
			})

			assert.NotNil(t, result)
			testutil.AssertDefinitionResult(t, result, tc.expected)

			if tc.expectedError == "" {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
