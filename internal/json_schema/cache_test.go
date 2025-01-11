package jsonschema

import (
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
)

func TestCreateNewSchema(t *testing.T) {
	callCount := 0

	sut := &JSONSchemaCache{
		cache: make(map[uri.URI]cachedGeneratedJSONSchema),
		schemaCreation: func(chart *charts.Chart) (string, error) {
			callCount++
			return `filepath`, nil
		},
	}

	chart := &charts.Chart{
		ValuesFiles: &charts.ValuesFiles{
			MainValuesFile: &charts.ValuesFile{
				Values: map[string]interface{}{
					"key": "value",
				},
			},
			OverlayValuesFile: &charts.ValuesFile{
				Values: map[string]interface{}{
					"other": "value",
				},
			},
			AdditionalValuesFiles: []*charts.ValuesFile{},
		},
		ChartMetadata: &charts.ChartMetadata{},
		RootURI:       "chart0",
	}
	result, err := sut.GetJsonSchemaForChart(chart)

	assert.NoError(t, err)
	assert.Equal(t, `filepath`, result)

	result2, err := sut.GetJsonSchemaForChart(chart)

	assert.NoError(t, err)
	assert.Equal(t, `filepath`, result2)
	assert.Equal(t, 1, callCount)

	chart.ValuesFiles.MainValuesFile.Values["key"] = "value2"
	result3, err := sut.GetJsonSchemaForChart(chart)
	assert.NoError(t, err)
	assert.Equal(t, `filepath`, result3)
	assert.Equal(t, 2, callCount)

	otherChart := &charts.Chart{
		ValuesFiles: &charts.ValuesFiles{
			MainValuesFile: &charts.ValuesFile{
				Values: map[string]interface{}{
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
	assert.Equal(t, `filepath`, result4)
	assert.Equal(t, 3, callCount)
}
