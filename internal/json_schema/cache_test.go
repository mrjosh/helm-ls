package jsonschema

import (
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
	"helm.sh/helm/v3/pkg/chart"
)

func TestCreateNewSchema(t *testing.T) {
	callCount := 0

	sut := &JSONSchemaCache{
		cache: make(map[uri.URI]cachedGeneratedJSONSchema),
		schemaCreation: func(chart *charts.Chart, chartStore *charts.ChartStore, getSchemaPathForChart func(chart *charts.Chart) string) (GeneratedChartJSONSchema, error) {
			callCount++
			return GeneratedChartJSONSchema{
				schema:       &Schema{},
				dependencies: []*charts.Chart{},
			}, nil
		},
	}

	testChart := &charts.Chart{
		HelmChart: &chart.Chart{},
		ValuesFiles: &charts.ValuesFiles{
			MainValuesFile: &charts.ValuesFile{
				Values: map[string]any{
					"key": "value",
				},
			},
			OverlayValuesFile: &charts.ValuesFile{
				Values: map[string]any{
					"other": "value",
				},
			},
			AdditionalValuesFiles: []*charts.ValuesFile{},
		},
		ChartMetadata: &charts.ChartMetadata{},
		RootURI:       "chart0",
	}
	result, err := sut.GetJsonSchemaForChart(testChart)
	expectedPath := "403899339-.json"

	assert.NoError(t, err)
	assert.Equal(t, expectedPath, result)

	result2, err := sut.GetJsonSchemaForChart(testChart)

	assert.NoError(t, err)
	assert.Equal(t, expectedPath, result2)
	assert.Equal(t, 1, callCount)

	testChart.ValuesFiles.MainValuesFile.Values["key"] = "value2"
	result3, err := sut.GetJsonSchemaForChart(testChart)
	assert.NoError(t, err)
	assert.Equal(t, "473433085-.json", result3)
	assert.Equal(t, 2, callCount)

	otherChart := &charts.Chart{
		HelmChart: &chart.Chart{},
		ValuesFiles: &charts.ValuesFiles{
			MainValuesFile: &charts.ValuesFile{
				Values: map[string]any{
					"key": "value",
				},
			},
		},
		ChartMetadata: &charts.ChartMetadata{},
		RootURI:       "chart1",
		ParentChart:   charts.ParentChart{},
	}
	result4, err := sut.GetJsonSchemaForChart(otherChart)
	assert.NoError(t, err)
	assert.Equal(t, "403899339-.json", result4)
	assert.Equal(t, 3, callCount)
}
