package jsonschema

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
	"helm.sh/helm/v3/pkg/chart"
)

func TestCreateNewSchema(t *testing.T) {
	callCount := 0
	tempDir := t.TempDir()

	sut := &JSONSchemaCache{
		cache: make(map[uri.URI]cachedGeneratedJSONSchema),
		schemaCreation: func(chart *charts.Chart, chartStore *charts.ChartStore, getSchemaPathForChart func(chart *charts.Chart) string) (GeneratedChartJSONSchema, error) {
			callCount++
			return GeneratedChartJSONSchema{
				schema:       &Schema{},
				dependencies: []*charts.Chart{},
			}, nil
		},
		schemaFilesDir: tempDir,
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
		RootURI:       uri.File(filepath.Join(tempDir, "chart0")),
	}
	result, err := sut.GetJSONSchemaForChart(testChart)
	expectedPath := filepath.Join(tempDir, "403899339-chart0.json")

	assert.NoError(t, err)
	assert.Equal(t, expectedPath, result)

	result2, err := sut.GetJSONSchemaForChart(testChart)

	assert.NoError(t, err)
	assert.Equal(t, expectedPath, result2)
	assert.Equal(t, 1, callCount)

	testChart.ValuesFiles.MainValuesFile.Values["key"] = "value2"
	result3, err := sut.GetJSONSchemaForChart(testChart)
	assert.NoError(t, err)
	expectedPath = filepath.Join(tempDir, "473433085-chart0.json")
	assert.Equal(t, expectedPath, result3)
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
		RootURI:       uri.File(filepath.Join(tempDir, "chart1")),
		ParentChart:   charts.ParentChart{},
	}
	result4, err := sut.GetJSONSchemaForChart(otherChart)
	assert.NoError(t, err)
	expectedPath = filepath.Join(tempDir, "403899339-chart1.json")
	assert.Equal(t, expectedPath, result4)
	assert.Equal(t, 3, callCount)
}

func TestDependecyProcessing(t *testing.T) {
	timeout := time.After(2 * time.Second)
	done := make(chan bool)
	go func() {
		callCount := 0
		tempDir := t.TempDir()

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
			RootURI:       uri.File(filepath.Join(tempDir, "chart0")),
		}

		testChartDep := &charts.Chart{
			HelmChart: &chart.Chart{},
			ValuesFiles: &charts.ValuesFiles{
				MainValuesFile: &charts.ValuesFile{
					Values: map[string]any{
						"key": "value",
					},
				},
				AdditionalValuesFiles: []*charts.ValuesFile{},
			},
			ChartMetadata: &charts.ChartMetadata{},
			RootURI:       uri.File(filepath.Join(tempDir, "chart1")),
		}

		sut := &JSONSchemaCache{
			cache: make(map[uri.URI]cachedGeneratedJSONSchema),
			schemaCreation: func(chart *charts.Chart, chartStore *charts.ChartStore, getSchemaPathForChart func(chart *charts.Chart) string) (GeneratedChartJSONSchema, error) {
				callCount++
				if chart == testChart {
					return GeneratedChartJSONSchema{
						schema: &Schema{},
						dependencies: []*charts.Chart{
							testChartDep,
						},
					}, nil
				}
				if chart == testChartDep {
					return GeneratedChartJSONSchema{
						schema: &Schema{},
						dependencies: []*charts.Chart{
							testChart,
						},
					}, nil
				}
				return GeneratedChartJSONSchema{
					schema:       &Schema{},
					dependencies: []*charts.Chart{},
				}, nil
			},
			schemaFilesDir: tempDir,
		}

		result, err := sut.GetJSONSchemaForChart(testChart)
		expectedPath := filepath.Join(tempDir, "403899339-chart0.json")
		assert.NoError(t, err)
		assert.Equal(t, expectedPath, result)
		assert.Equal(t, 2, callCount)
		done <- true
	}()

	select {
	case <-timeout:
		t.Fatal("Test didn't finish in time")
	case <-done:
	}
}
