package jsonschema

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
)

var rootUri = uri.File("../../testdata/dependenciesExample/")

func TestCreateJsonSchemaForChart(t *testing.T) {
	addChartCallback := func(chart *charts.Chart) {}
	chartStore := charts.NewChartStore(rootUri, charts.NewChart, addChartCallback)
	chart, err := chartStore.GetChartForURI(rootUri)

	assert.NoError(t, err)
	schemaFile, err := createJsonSchemaForChart(chart, chartStore)

	content, err := os.ReadFile(schemaFile)
	assert.NoError(t, err)

	schema := &Schema{}
	json.Unmarshal(content, schema)

	assert.NoError(t, err)
	assert.NotNil(t, schema)

	assert.Contains(t, string(content), "onlyInSubchartValues", "Schema should contain value from subchart")
	content, _ = json.MarshalIndent(schema, "", "  ")
	println(string(content))
	t.Fail()

	assert.Contains(t, string(content), "targetCPUUtilizationPercentage")
	assert.Contains(t, string(content), "subchartWithoutGlobal")
}

func TestCreateJsonSchemaForChartNested(t *testing.T) {
	rootUri := uri.File("../../testdata/nestedDependenciesExample/charts/onceNested/")
	addChartCallback := func(chart *charts.Chart) {}
	chartStore := charts.NewChartStore(rootUri, charts.NewChart, addChartCallback)
	chart, err := chartStore.GetChartForURI(rootUri)

	assert.NoError(t, err)
	schemaFile, err := createJsonSchemaForChart(chart, chartStore)

	content, err := os.ReadFile(schemaFile)
	assert.NoError(t, err)

	schema := &Schema{}
	json.Unmarshal(content, schema)

	assert.NoError(t, err)
	assert.NotNil(t, schema)

	assert.Contains(t, string(content), "onlyInSubchartValues", "Schema should contain value from subchart")
	content, _ = json.MarshalIndent(schema, "", "  ")
	println(string(content))
	t.Fail()

	assert.Contains(t, string(content), "targetCPUUtilizationPercentage")
	assert.Contains(t, string(content), "subchartWithoutGlobal")
}
