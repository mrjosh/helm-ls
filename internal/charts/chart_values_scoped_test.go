package charts

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
)

func TestGetScopedValuesFiles(t *testing.T) {
	tests := []struct {
		name         string
		chartRootDir string
		expected     []*ScopedValuesFiles
	}{
		{
			name:         "Test with dependenciesExample",
			chartRootDir: "../../testdata/dependenciesExample/",
			expected: []*ScopedValuesFiles{
				{Scope: []string{}, SubScope: []string{}, ValuesFiles: nil},
				{Scope: []string{"subchartexample"}, SubScope: []string{}, ValuesFiles: nil},
				{Scope: []string{"common"}, SubScope: []string{}, ValuesFiles: nil},
			},
		},
		{
			name:         "Test with dependenciesExample",
			chartRootDir: "../../testdata/dependenciesExample/charts/subchartexample/",
			expected: []*ScopedValuesFiles{
				{Scope: []string{}, SubScope: []string{}, ValuesFiles: nil},
				{Scope: []string{}, SubScope: []string{"subchartexample"}, ValuesFiles: nil},
			},
		},
		{
			name:         "Test with nestedDependenciesExample for child",
			chartRootDir: "../../testdata/nestedDependenciesExample/charts/onceNested/charts/twiceNested/",
			expected: []*ScopedValuesFiles{
				{Scope: []string{}, SubScope: []string{}, ValuesFiles: nil},
				{Scope: []string{}, SubScope: []string{"twiceNested"}, ValuesFiles: nil},
				{Scope: []string{}, SubScope: []string{"onceNested", "twiceNested"}, ValuesFiles: nil},
			},
		},
		{
			name:         "Test with nestedDependenciesExample for parent",
			chartRootDir: "../../testdata/nestedDependenciesExample/",
			expected: []*ScopedValuesFiles{
				{Scope: []string{}, SubScope: []string{}, ValuesFiles: nil},
				{Scope: []string{"onceNested"}, SubScope: []string{}, ValuesFiles: nil},
				{Scope: []string{"nestedDependenciesExample", "subchartexample"}, SubScope: []string{}, ValuesFiles: nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addChartCallback := func(chart *Chart) {}
			chartStore := NewChartStore(uri.File(tt.chartRootDir), NewChart, addChartCallback)
			chart, err := chartStore.GetChartForURI(uri.File(tt.chartRootDir))

			assert.NoError(t, err)

			result := chart.GetScopedValuesFiles(chartStore)

			assert.Len(t, result, len(tt.expected))

			assert.True(t, slices.ContainsFunc(result,
				func(actual *ScopedValuesFiles) bool {
					return slices.ContainsFunc(tt.expected, func(expected *ScopedValuesFiles) bool {
						return slices.Equal(actual.Scope, expected.Scope)
					})
				},
			))
		})
	}
}
