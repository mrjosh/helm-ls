package jsonschema

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
)

func TestSchemGenerationSnapshot(t *testing.T) {
	snapshotTest(t, "../../testdata/dependenciesExample/")
	snapshotTest(t, "../../testdata/dependenciesExample/charts/subchartexample/")
	snapshotTest(t, "../../testdata/nestedDependenciesExample/")
	snapshotTest(t, "../../testdata/nestedDependenciesExample/charts/onceNested/")
	snapshotTest(t, "../../testdata/nestedDependenciesExample/charts/onceNested/charts/twiceNested/")
}

func snapshotTest(t *testing.T, path string) {
	schema := getSchemaForChart(t, uri.File(path))
	snaps.MatchSnapshot(t, schemaToJSON(schema))
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
	rootUri := uri.File("../../testdata/dependenciesExample/charts/subchartexample/")
	schema := getSchemaForChart(t, rootUri)

	definitionsDoesContainPropertyGlobalProperty(t, schema, []string{"subchart"})
	expectedRef := &Schema{Ref: "#/$defs/global"}
	refsContains(t, schema, expectedRef)
}

func TestHasParentValuesInSchema(t *testing.T) {
	rootUri := uri.File("../../testdata/dependenciesExample/charts/subchartexample/")
	schema := getSchemaForChart(t, rootUri)

	definitionsDoesContainProperty(t, schema, "dependeciesExample", []string{"fromParent"})
	expectedRef := &Schema{Ref: "#/$defs/dependeciesExample"}
	refsContains(t, schema, expectedRef)
}

func TestHasParentValuesForNestedInSchema(t *testing.T) {
	rootUri := uri.File("../../testdata/nestedDependenciesExample/charts/onceNested/")
	schema := getSchemaForChart(t, rootUri)

	definitionsDoesContainProperty(t, schema, "nestedDependenciesExample", []string{"fromRootForOnceNested"})
}

func TestHasParentValuesForTwiceNestedInSchema(t *testing.T) {
	rootUri := uri.File("../../testdata/nestedDependenciesExample/charts/onceNested/charts/twiceNested/")
	schema := getSchemaForChart(t, rootUri)

	definitionsDoesContainProperty(t, schema, "nestedDependenciesExample", []string{"twiceNested", "fromRootForOnceNested"})
	definitionsDoesContainProperty(t, schema, "onceNested", []string{"fromOnceNestedForTwiceNested"})
}

func TestHasValuesFromSubChartNestedInSchema(t *testing.T) {
	rootUri := uri.File("../../testdata/nestedDependenciesExample/charts/onceNested/")
	schema := getSchemaForChart(t, rootUri)
	definitionsDoesContainProperty(t, schema, "twiceNested", []string{"twiceNested", "onlyInTwiceNested"})
}

func definitionsDoesContainProperty(t *testing.T, schema *Schema, definitionName string, propertyPath []string) {
	definitionsDoesContainPropertyGeneric(t, schema, definitionName, []string{}, propertyPath)
}

func definitionsDoesContainPropertyGlobalProperty(t *testing.T, schema *Schema, propertyPath []string) {
	definitionsDoesContainPropertyGeneric(t, schema, "global", []string{"global"}, propertyPath)
}

func definitionsDoesContainPropertyGeneric(t *testing.T, schema *Schema, definitionName string, prePropertyPath []string, propertyPath []string) {
	subSchema := schema.Definitions[definitionName]
	assert.NotNil(t, subSchema, "Definition %s should exist on schema, but does not, schema: %s", definitionName, schemaToJSON(schema))

	found, ok := false, true

	for _, preProperty := range prePropertyPath {
		props := subSchema.Properties
		subSchema, ok = props[preProperty]
		if !ok {
			t.Fatalf("Definition %s should contain preProperty %s, but does not, schema: %s", definitionName, preProperty, schemaToJSON(schema.Definitions[definitionName]))
		}
	}

	assert.NotNil(t, subSchema.AllOf, "Subschema has no AllOf property: %s", schemaToJSON(subSchema))
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
