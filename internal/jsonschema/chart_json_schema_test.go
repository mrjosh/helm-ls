package jsonschema

import (
	"encoding/json"
	"fmt"
	"runtime"
	"slices"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
)

func TestSchemGenerationSnapshot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on windows because snapshots are not platform independent")
	}
	snapshotTest(t, "../../testdata/dependenciesExample/")
	snapshotTest(t, "../../testdata/dependenciesExample/charts/subchartexample/")
	snapshotTest(t, "../../testdata/nestedDependenciesExample/")
	snapshotTest(t, "../../testdata/nestedDependenciesExample/charts/onceNested/")
	snapshotTest(t, "../../testdata/nestedDependenciesExample/charts/onceNested/charts/twiceNested/")
}

func snapshotTest(t *testing.T, path string) {
	schema, _ := getSchemaForChart(t, uri.File(path))
	snaps.MatchStandaloneJSON(t, schema)
}

func TestHasOtherValueFilesInSameChartInSchema(t *testing.T) {
	rootURI := uri.File("../../testdata/dependenciesExample/")
	schema, _ := getSchemaForChart(t, rootURI)

	definitionsDoesContainProperty(t, schema, "dependenciesExample", []string{"autoscaling", "targetCPUUtilizationPercentage"})
	definitionsDoesContainProperty(t, schema, "dependenciesExample", []string{"fromTheFileA"})
	definitionsDoesContainProperty(t, schema, "dependenciesExample", []string{"fromTheFileB"})

	expectedRef := &Schema{Ref: "#/$defs/dependenciesExample"}
	refsContains(t, schema, expectedRef)
}

func TestPointsToValuesFromSubChart(t *testing.T) {
	schema, _ := getSchemaForChart(t, uri.File("../../testdata/dependenciesExample/"))
	subchartexampleSchema, path := getSchemaForChart(t, uri.File("../../testdata/dependenciesExample/charts/subchartexample/"))

	definitionsDoesContainProperty(t, subchartexampleSchema, "subchartexample", []string{"onlyInSubchartValues"})
	definitionsDoesContainProperty(t, subchartexampleSchema, "subchartexample", []string{"subchartWithoutGlobal"})
	expectedRef := &Schema{Ref: fmt.Sprintf("%s#/$defs/subchartexample", uri.File(path))}
	refsContainsNested(t, schema, expectedRef, []string{"subchartexample"})
}

func TestPointsToValuesValuesFromDependencySubChart(t *testing.T) {
	schema, _ := getSchemaForChart(t, uri.File("../../testdata/dependenciesExample/"))

	// TODO: this test does not check if the schema for the dependency is correct
	expectedRef := &Schema{Ref: fmt.Sprintf("%s#/$defs/common", uri.File("/common"))}
	refsContainsNested(t, schema, expectedRef, []string{"common"})
}

func TestHasGlobalValuesInSchema(t *testing.T) {
	schema, _ := getSchemaForChart(t, uri.File("../../testdata/dependenciesExample/charts/subchartexample/"))

	definitionsDoesContainPropertyGlobalProperty(t, schema, []string{"subchart"})
	expectedRef := &Schema{Properties: map[string]*Schema{"global": {Ref: "#/$defs/global"}}}
	refsContains(t, schema, expectedRef)
}

func TestHasParentValuesInSchema(t *testing.T) {
	schema, _ := getSchemaForChart(t, uri.File("../../testdata/dependenciesExample/charts/subchartexample/"))
	parentSchema, parentPath := getSchemaForChart(t, uri.File("../../testdata/dependenciesExample/"))

	definitionsDoesContainPropertyInAllOf(t, parentSchema, "dependenciesExample", []string{"subchartexample", "fromParent"})
	expectedRef := &Schema{Ref: fmt.Sprintf("%s#/$defs/dependenciesExample/allOf/0/properties/subchartexample", uri.File(parentPath))}
	refsContains(t, schema, expectedRef)
}

func TestHasParentValuesForNestedInSchema(t *testing.T) {
	schema, _ := getSchemaForChart(t, uri.File("../../testdata/nestedDependenciesExample/charts/onceNested/"))
	parentSchema, parentPath := getSchemaForChart(t, uri.File("../../testdata/nestedDependenciesExample"))

	definitionsDoesContainPropertyInAllOf(t, parentSchema, "nestedDependenciesExample", []string{"onceNested", "fromRootForOnceNested"})
	expectedRef := &Schema{Ref: fmt.Sprintf("%s#/$defs/nestedDependenciesExample/allOf/0/properties/onceNested", uri.File(parentPath))}
	refsContains(t, schema, expectedRef)
}

func TestHasParentValuesForTwiceNestedInSchema(t *testing.T) {
	schema, _ := getSchemaForChart(t, uri.File("../../testdata/nestedDependenciesExample/charts/onceNested/charts/twiceNested/"))
	assert.NotNil(t, schema)
	parentSchema, parentPath := getSchemaForChart(t, uri.File("../../testdata/nestedDependenciesExample/charts/onceNested/"))
	rootSchema, rootPath := getSchemaForChart(t, uri.File("../../testdata/nestedDependenciesExample"))

	definitionsDoesContainPropertyInAllOf(t, parentSchema, "onceNested", []string{"twiceNested", "fromOnceNestedForTwiceNested"})
	definitionsDoesContainPropertyInAllOf(t, rootSchema, "nestedDependenciesExample", []string{"onceNested", "twiceNested", "fromRootForTwiceNested"})

	expectedRef := &Schema{Ref: fmt.Sprintf("%s#/$defs/onceNested/allOf/0/properties/twiceNested", uri.File(parentPath))}
	refsContains(t, schema, expectedRef)
	expectedRef = &Schema{Ref: fmt.Sprintf("%s#/$defs/nestedDependenciesExample/allOf/0/properties/onceNested/properties/twiceNested", uri.File(rootPath))}
	refsContains(t, schema, expectedRef)
}

func TestHasValuesFromSubChartNestedInSchema(t *testing.T) {
	schema, _ := getSchemaForChart(t, uri.File("../../testdata/nestedDependenciesExample/charts/onceNested/"))
	subChart, subChartPath := getSchemaForChart(t, uri.File("../../testdata/nestedDependenciesExample/charts/onceNested/charts/twiceNested/"))
	definitionsDoesContainPropertyInAllOf(t, subChart, "twiceNested", []string{"onlyInTwiceNested"})

	expectedRef := &Schema{Ref: fmt.Sprintf("%s#/$defs/twiceNested", uri.File(subChartPath))}
	refsContainsNested(t, schema, expectedRef, []string{"twiceNested"})
}

func definitionsDoesContainProperty(t *testing.T, schema *Schema, definitionName string, propertyPath []string) {
	definitionsDoesContainPropertyGeneric(t, schema, definitionName, []string{}, propertyPath)
}

func definitionsDoesContainPropertyGlobalProperty(t *testing.T, schema *Schema, propertyPath []string) {
	definitionsDoesContainPropertyInAllOf(t, schema, "global", propertyPath)
}

func definitionsDoesContainPropertyInAllOf(t *testing.T, schema *Schema, definitionName string, propertyPath []string) {
	subSchema := schema.Definitions[definitionName]
	assert.NotNil(t, subSchema, "Definition %s should exist on schema, but does not, schema: %s", definitionName, schemaToJSON(schema))
	allOf := subSchema.AllOf
	assert.NotNil(t, allOf, "Subschema has no AllOf property: %s", schemaToJSON(subSchema))
	assert.Condition(t, func() bool {
		found := false
		for _, subSchema := range allOf {
			for _, property := range propertyPath {
				props := subSchema.Properties
				tmpSubSchema, ok := props[property]
				if !ok {
					found = false
					break
				}
				subSchema = tmpSubSchema
				found = true
			}
			if found {
				return true
			}
		}
		return false
	}, "Definition %s should contain property %s, but does not, schema: %s", definitionName, propertyPath, schemaToJSON(subSchema))
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
		return fmt.Sprintf("<error marshaling schema: %v>", err)
	}
	return string(schemaJSON)
}

func refsContainsNested(t *testing.T, schema *Schema, expectedRef *Schema, nesting []string) {
	slices.Reverse(nesting)
	nestedSchema := expectedRef

	for _, nested := range nesting {
		nestedSchema = &Schema{Properties: map[string]*Schema{
			nested: nestedSchema,
		}}
	}
	refsContains(t, schema, nestedSchema)
}

func refsContains(t *testing.T, schema *Schema, expectedRef *Schema) {
	expectedRefsJSON := schemaToJSON(expectedRef)

	allOfConverted := []string{}
	for _, schema := range schema.AllOf {
		allOfConverted = append(allOfConverted, schemaToJSON(schema))
	}
	assert.Contains(t, allOfConverted, expectedRefsJSON, "Schema should contain ref %s, but does not", expectedRefsJSON)
}

func getSchemaForChart(t *testing.T, rootURI uri.URI) (*Schema, string) {
	addChartCallback := func(chart *charts.Chart) {}
	chartStore := charts.NewChartStore(rootURI, charts.NewChart, addChartCallback)
	chart, err := chartStore.GetChartForURI(rootURI)

	getSchemaPathForChart := func(chart *charts.Chart) string {
		return "/" + chart.Name()
	}

	assert.NoError(t, err)
	generatedChartJSONSchema, err := CreateJSONSchemaForChart(chart, chartStore, getSchemaPathForChart)
	assert.NoError(t, err)

	return generatedChartJSONSchema.schema, getSchemaPathForChart(chart)
}
