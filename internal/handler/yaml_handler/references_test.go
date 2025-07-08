package yamlhandler

import (
	"context"
	"testing"

	"github.com/mrjosh/helm-ls/internal/testutil"
	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestReferences(t *testing.T) {
	testCases := []struct {
		desc       string
		filepath   string
		markedLine string

		expected      []testutil.ExpectedLocationsResult
		expectedError string
	}{
		{
			"Only defined in same file",
			"../../../testdata/example/values.yaml",
			"replica^Count: 1",
			[]testutil.ExpectedLocationsResult{
				{
					Filepath:   "../../../testdata/example/values.yaml",
					MarkedLine: "§replicaCount§: 1",
				},
			},
			"",
		},
		{
			"Defined in multiple files in same chart",
			"../../../testdata/dependenciesExample/values.a.yaml",
			"ima^ge:",
			[]testutil.ExpectedLocationsResult{
				{
					Filepath:   "../../../testdata/dependenciesExample/values.b.yaml",
					MarkedLine: "§image§:",
				},
				{
					Filepath:   "../../../testdata/dependenciesExample/values.yaml",
					MarkedLine: "§image§:",
				},
				{
					Filepath:   "../../../testdata/dependenciesExample/values.a.yaml",
					MarkedLine: "§image§:",
				},
			},
			"",
		},
		{
			"From parent to subchart",
			"../../../testdata/dependenciesExample/values.yaml",
			"subchartWithout^Global: worksToo",

			[]testutil.ExpectedLocationsResult{
				{
					Filepath:   "../../../testdata/dependenciesExample/charts/subchartexample/values.yaml",
					MarkedLine: "§subchartWithoutGlobal§: works",
				},
				{
					Filepath:   "../../../testdata/dependenciesExample/values.yaml",
					MarkedLine: "§subchartWithoutGlobal§: worksToo",
				},
			},
			"",
		},
		{
			"From parent to subchart global",
			"../../../testdata/dependenciesExample/values.yaml",
			"  subch^art: works",

			[]testutil.ExpectedLocationsResult{
				{
					Filepath:   "../../../testdata/dependenciesExample/charts/subchartexample/values.yaml",
					MarkedLine: "  §subchart§: works",
				},
				{
					Filepath:   "../../../testdata/dependenciesExample/values.yaml",
					MarkedLine: "  §subchart§: works",
				},
			},
			"",
		},
		{
			"From subchart to parent",
			"../../../testdata/dependenciesExample/charts/subchartexample/values.yaml",
			"^subchartWithoutGlobal: works",

			[]testutil.ExpectedLocationsResult{
				{
					Filepath:   "../../../testdata/dependenciesExample/values.yaml",
					MarkedLine: "§subchartWithoutGlobal§: worksToo",
				},
				{
					Filepath:   "../../../testdata/dependenciesExample/charts/subchartexample/values.yaml",
					MarkedLine: "§subchartWithoutGlobal§: works",
				},
			},
			"",
		},
		{
			"From subchart to parent multiple files",
			"../../../testdata/dependenciesExample/charts/subchartexample/values.yaml",
			"^global:",

			[]testutil.ExpectedLocationsResult{
				{
					Filepath:   "../../../testdata/dependenciesExample/values.yaml",
					MarkedLine: "§global§:",
				},
				{
					Filepath:   "../../../testdata/dependenciesExample/values.a.yaml",
					MarkedLine: "§global§:",
				},
				{
					Filepath:   "../../../testdata/dependenciesExample/charts/subchartexample/values.yaml",
					MarkedLine: "§global§:",
				},
			},
			"",
		},
		{
			"From parent to subchart values when on subchart name",
			"../../../testdata/dependenciesExample//values.yaml",
			"^subchartexample:",

			[]testutil.ExpectedLocationsResult{
				{
					Filepath:   "../../../testdata/dependenciesExample/charts/subchartexample/values.yaml",
					MarkedLine: "§§global:",
				},
				{
					Filepath:   "../../../testdata/dependenciesExample/values.b.yaml",
					MarkedLine: "§subchartexample§:",
				},
				{
					Filepath:   "../../../testdata/dependenciesExample/values.yaml",
					MarkedLine: "§subchartexample§:",
				},
			},
			"",
		},
		{
			"From chart to parent and child",
			"../../../testdata/nestedDependenciesExample/charts/onceNested/values.yaml",
			"^twiceNested:",

			[]testutil.ExpectedLocationsResult{
				{
					Filepath:   "../../../testdata/nestedDependenciesExample/charts/onceNested/charts/twiceNested/values.yaml",
					MarkedLine: "§§replicaCount: ",
				},
				{
					Filepath:   "../../../testdata/nestedDependenciesExample/values.yaml",
					MarkedLine: "  §twiceNested§:",
				},
				{
					Filepath:   "../../../testdata/nestedDependenciesExample/charts/onceNested/values.yaml",
					MarkedLine: "§twiceNested§:",
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

			result, err := h.References(context.Background(), &lsp.ReferenceParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{
						URI: uri.File(tc.filepath),
					},
					Position: pos,
				},
			})

			assert.NotNil(t, result)
			testutil.AssertLocationsResult(t, result, tc.expected)

			if tc.expectedError == "" {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
