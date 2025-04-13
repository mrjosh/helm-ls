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
	assert := assert.New(t)
	addChartCallback := func(chart *charts.Chart) {}
	chartStore := charts.NewChartStore(rootUri, charts.NewChart, addChartCallback)
	chart, err := chartStore.GetChartForURI(rootUri)

	assert.NoError(err)
	schemaFile, err := createJsonSchemaForChart(chart, chartStore)

	content, err := os.ReadFile(schemaFile)
	assert.NoError(err)

	schema := &Schema{}
	json.Unmarshal(content, schema)

	assert.NoError(err)
	assert.NotNil(schema)

	subchartSchemaDef := schema.Definitions["subchartexample"]
	assert.NotNil(subchartSchemaDef)
	properties := subchartSchemaDef.AllOf[0].Properties["subchartexample"]
	assert.NotNil(properties)
	assert.NotNil(properties.Properties["onlyInSubchartValues"])

	expectedRefs := &Schema{Ref: "#/$defs/subchartexample"}
	expectedRefsJSON, err := json.Marshal(expectedRefs)
	assert.NoError(err)

	allOfConverted := []string{}
	for _, schema := range schema.AllOf {
		schemaJSON, err := json.Marshal(schema)
		assert.NoError(err)
		allOfConverted = append(allOfConverted, string(schemaJSON))
	}
	assert.Contains(allOfConverted, string(expectedRefsJSON))

	assert.Contains(string(content), "onlyInSubchartValues", "Schema should contain value from subchart")
	content, _ = json.MarshalIndent(schema, "", "  ")
	println(schemaFile)

	assert.Contains(string(content), "targetCPUUtilizationPercentage")
	assert.Contains(string(content), "subchartWithoutGlobal")
	// t.Fail()
}

func TestCreateJsonSchemaForChartNested(t *testing.T) {
	t.Skip()
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
	// t.Fail()

	assert.Contains(t, string(content), "targetCPUUtilizationPercentage")
	assert.Contains(t, string(content), "subchartWithoutGlobal")
}
