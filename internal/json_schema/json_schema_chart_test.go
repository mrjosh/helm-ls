package jsonschema

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
)

func getSchemaForChart(t *testing.T, rootUri uri.URI) *Schema {
	addChartCallback := func(chart *charts.Chart) {}
	chartStore := charts.NewChartStore(rootUri, charts.NewChart, addChartCallback)
	chart, err := chartStore.GetChartForURI(rootUri)

	assert.NoError(t, err)
	schemaFile, err := createJsonSchemaForChart(chart, chartStore)
	println(schemaFile)
	assert.NoError(t, err)

	content, err := os.ReadFile(schemaFile)
	assert.NoError(t, err)

	schema := &Schema{}
	json.Unmarshal(content, schema)

	return schema
}

func TestHasOtherValueFilesInSameChartInSchema(t *testing.T) {
	rootUri := uri.File("../../testdata/dependenciesExample/")
	schema := getSchemaForChart(t, rootUri)

	definitionsDoesContainProperty(t, schema, "dependeciesExample", []string{"autoscaling", "targetCPUUtilizationPercentage"})
	definitionsDoesContainProperty(t, schema, "dependeciesExample", []string{"fromTheFileA"})
	definitionsDoesContainProperty(t, schema, "dependeciesExample", []string{"fromTheFileB"})

	expectedRef := &Schema{Ref: "#/$defs/dependeciesExample"}
	refsContains(t, schema, expectedRef)
}

func TestHasValuesFromSubChartInSchema(t *testing.T) {
	rootUri := uri.File("../../testdata/dependenciesExample/")
	schema := getSchemaForChart(t, rootUri)

	definitionsDoesContainProperty(t, schema, "subchartexample", []string{"subchartexample", "onlyInSubchartValues"})
	definitionsDoesContainProperty(t, schema, "subchartexample", []string{"subchartexample", "subchartWithoutGlobal"})
	expectedRef := &Schema{Ref: "#/$defs/subchartexample"}
	refsContains(t, schema, expectedRef)
}

func TestHasValuesFromDependencySubChartInSchema(t *testing.T) {
	rootUri := uri.File("../../testdata/dependenciesExample/")
	schema := getSchemaForChart(t, rootUri)

	definitionsDoesContainProperty(t, schema, "common", []string{"common", "exampleValue"})
	expectedRef := &Schema{Ref: "#/$defs/common"}
	refsContains(t, schema, expectedRef)
}

func TestHasGlobalValuesInSchema(t *testing.T) {
	rootUri := uri.File("../../testdata/dependenciesExample/")
	schema := getSchemaForChart(t, rootUri)
	definitionsDoesContainProperty(t, schema, "global", []string{"global", "subchart"})
}

func TestCreateJsonSchemaForChart(t *testing.T) {
	rootUri := uri.File("../../testdata/dependenciesExample/")

	schema := getSchemaForChart(t, rootUri)

	definitionsDoesContainProperty(t, schema, "subchartexample", []string{"subchartexample", "onlyInSubchartValues"})
	definitionsDoesContainProperty(t, schema, "subchartexample", []string{"subchartexample", "subchartWithoutGlobal"})
	expectedRef := &Schema{Ref: "#/$defs/subchartexample"}
	refsContains(t, schema, expectedRef)

	definitionsDoesContainProperty(t, schema, "dependeciesExample", []string{"autoscaling", "targetCPUUtilizationPercentage"})
	definitionsDoesContainProperty(t, schema, "common", []string{"common", "exampleValue"})

	expectedRef = &Schema{Ref: "#/$defs/dependeciesExample"}
	refsContains(t, schema, expectedRef)

	expectedRef = &Schema{Ref: "#/$defs/common"}
	refsContains(t, schema, expectedRef)
}

func definitionsDoesContainProperty(t *testing.T, schema *Schema, definitionName string, propertyPath []string) {
	subSchema := schema.Definitions[definitionName]
	assert.NotNil(t, subSchema, "Definition %s should exist on schema, but does not", definitionName)

	found, ok := false, true

	assert.Condition(t, func() bool {
		for _, subSchema := range subSchema.AllOf {
			for _, property := range propertyPath {
				props := subSchema.Properties
				subSchema, ok = props[property]
				if !ok {
					found = false
					break
				}
				found = true
			}

			if subSchema != nil {
				found = true
				return found
			}
		}
		return found
	}, "Definition %s should contain property %s, but does not, schema: %s", definitionName, propertyPath, schemaToJSON(schema.Definitions[definitionName]))
}

func schemaToJSON(schema *Schema) string {
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		panic(err)
	}
	return string(schemaJSON)
}

func refsContains(t *testing.T, schema *Schema, expectedRef *Schema) {
	expectedRefsJSON := schemaToJSON(expectedRef)

	allOfConverted := []string{}
	for _, schema := range schema.AllOf {
		allOfConverted = append(allOfConverted, schemaToJSON(schema))
	}
	assert.Contains(t, allOfConverted, expectedRefsJSON, "Schema should contain ref %s, but does not", expectedRefsJSON)
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

	definitionsDoesContainProperty(t, schema, "twiceNested", []string{"twiceNested", "onlyInTwiceNested"})
	definitionsDoesContainProperty(t, schema, "nestedDependenciesExample", []string{"fromRootForOnceNested"})
	println(schemaFile)
}
