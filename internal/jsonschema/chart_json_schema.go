package jsonschema

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/log"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/uri"
)

var logger = log.GetLogger()

// SchemaGenerator handles the generation of JSON schemas for Helm charts
type SchemaGenerator struct {
	chart                 *charts.Chart
	chartStore            *charts.ChartStore
	getSchemaPathForChart func(chart *charts.Chart) string
	allOf                 []*Schema
	globalSchemas         []*Schema
	definitions           map[string]*Schema
	errors                []error
}

// NewSchemaGenerator creates a new SchemaGenerator instance
func NewSchemaGenerator(chart *charts.Chart, chartStore *charts.ChartStore, getSchemaPathForChart func(chart *charts.Chart) string) *SchemaGenerator {
	return &SchemaGenerator{
		chart:                 chart,
		chartStore:            chartStore,
		getSchemaPathForChart: getSchemaPathForChart,
		allOf:                 []*Schema{},
		globalSchemas:         []*Schema{},
		definitions:           map[string]*Schema{},
		errors:                []error{},
	}
}

type GeneratedChartJSONSchema struct {
	schema       *Schema
	dependencies []*charts.Chart
}

func (g *SchemaGenerator) Generate() (GeneratedChartJSONSchema, error) {
	dependencies := []*charts.Chart{}

	// Process all scoped values files
	for _, scopedValuesfiles := range g.chart.GetScopedValuesFiles(g.chartStore) {
		if len(scopedValuesfiles.Scope) == 0 && len(scopedValuesfiles.SubScope) == 0 {
			g.generateSchemaForCurrentChart(scopedValuesfiles)
		} else {
			dependencies = append(dependencies, scopedValuesfiles.Chart)
			g.generateSchemaForRelatedChart(scopedValuesfiles)
		}
	}

	g.addGlobalDef()

	schema := generateSchemaWithAllOf(g.definitions, g.allOf)

	return GeneratedChartJSONSchema{
		schema:       schema,
		dependencies: dependencies,
	}, nil // TODO: collect errors
}

func (g *SchemaGenerator) generateSchemaForCurrentChart(scopedValuesfiles *charts.ScopedValuesFiles) {
	valuesSchemas := []*Schema{}
	for _, valuesFile := range scopedValuesfiles.ValuesFiles.AllValuesFiles() {
		subVals := g.processGlobalValsForCurrentChart(valuesFile)

		schema := generateJSONSchema(subVals, g.getDescriptionForValuesSchema(valuesFile))

		valuesSchemas = append(valuesSchemas, schema)
	}

	if schemaFileSchema := getSchemaFileSchema(scopedValuesfiles.Chart); schemaFileSchema != nil {
		valuesSchemas = append(valuesSchemas, schemaFileSchema)
	}

	g.addCurrentChartDef(scopedValuesfiles.Chart, valuesSchemas)
}

func (g *SchemaGenerator) generateSchemaForRelatedChart(scopedValuesfiles *charts.ScopedValuesFiles) {
	schemFilePath := uri.File(g.getSchemaPathForChart(scopedValuesfiles.Chart))

	refPointers := []string{""}

	if len(scopedValuesfiles.SubScope) > 0 {
		refPointers = []string{}
		refPointer := ""
		for _, v := range scopedValuesfiles.SubScope {
			refPointer = fmt.Sprintf("%s/properties/%s", refPointer, v)
		}

		for i, valuesFile := range scopedValuesfiles.ValuesFiles.AllValuesFiles() {
			vals := valuesFile.Values.AsMap()
			_, err := util.GetSubValuesForSelector(vals, scopedValuesfiles.SubScope)
			if err == nil {
				refPointers = append(refPointers, fmt.Sprintf("/allOf/%d%s", i, refPointer))
			}
		}
	}

	for _, refPointer := range refPointers {
		ref := fmt.Sprintf("%s#/$defs/%s%s",
			schemFilePath,
			scopedValuesfiles.Chart.Name(),
			refPointer,
		)

		schema := &Schema{Ref: ref}
		schema = nestSchemaInScopes(schema, scopedValuesfiles.Scope)
		g.allOf = append(g.allOf, schema)
	}

	g.addGlobalRef(scopedValuesfiles, schemFilePath)
}

// Will add a reference to the global schema if any global values are found
func (g *SchemaGenerator) addGlobalRef(scopedValuesfiles *charts.ScopedValuesFiles, schemFilePath uri.URI) {
	for _, valuesFile := range scopedValuesfiles.ValuesFiles.AllValuesFiles() {
		vals := valuesFile.Values.AsMap()
		_, err := util.GetSubValuesForSelector(vals, []string{"global"})
		if err == nil {
			globalSchemaRef := fmt.Sprintf("%s#/$defs/%s",
				schemFilePath,
				"global",
			)
			globalSchema := &Schema{Ref: globalSchemaRef}
			g.allOf = append(g.allOf, nestSchemaInScopes(globalSchema, []string{"global"}))
			break
		}
	}
}

// TODO: think about good descriptions
func (g *SchemaGenerator) getDescriptionForValuesSchema(valuesFile *charts.ValuesFile) string {
	return fmt.Sprintf("%s values from the file %s", g.chart.Name(), filepath.Base(valuesFile.URI.Filename()))
}

func (g *SchemaGenerator) getDescriptionForGlobalValues(valuesFile *charts.ValuesFile) string {
	return fmt.Sprintf("global values from the file %s", filepath.Base(valuesFile.URI.Filename()))
}

func (g *SchemaGenerator) processGlobalValsForCurrentChart(valuesFile *charts.ValuesFile) map[string]any {
	subVals := valuesFile.Values.AsMap()

	globalVals, ok := subVals["global"]
	if !ok {
		return subVals
	}

	globalValsMap, ok := globalVals.(map[string]any)
	if !ok {
		return subVals
	}

	globalSchema := generateJSONSchema(globalValsMap, g.getDescriptionForGlobalValues(valuesFile))
	g.globalSchemas = append(g.globalSchemas, globalSchema)

	subValsTmp := map[string]any{}
	for k, v := range subVals {
		if k != "global" {
			subValsTmp[k] = v
		}
	}
	return subValsTmp
}

func (g *SchemaGenerator) addCurrentChartDef(chart *charts.Chart, valuesSchemas []*Schema) {
	g.addDef(chart.Name(), &Schema{AllOf: valuesSchemas})
}

func (g *SchemaGenerator) addGlobalDef() {
	if len(g.globalSchemas) == 0 {
		return
	}

	g.addNestedDef("global", &Schema{AllOf: g.globalSchemas}, []string{"global"})
}

func (g *SchemaGenerator) addDef(name string, schema *Schema) {
	g.definitions[name] = schema
	g.allOf = append(g.allOf, &Schema{
		Ref: fmt.Sprintf("#/$defs/%s", name),
	})
}

func (g *SchemaGenerator) addNestedDef(name string, schema *Schema, nesting []string) {
	g.definitions[name] = schema
	g.allOf = append(g.allOf, nestSchemaInScopes(
		&Schema{
			Ref: fmt.Sprintf("#/$defs/%s", name),
		},
		nesting,
	),
	)
}

// Gets the schema from the values.schema.json file if the chart has one
func getSchemaFileSchema(chart *charts.Chart) *Schema {
	if chart.HelmChart.Schema != nil {
		schemaFileSchema := &Schema{}
		err := json.Unmarshal(chart.HelmChart.Schema, schemaFileSchema)
		if err != nil {
			logger.Error("Failed to unmarshal schema from helm chart "+chart.RootURI, err)
		} else {
			return schemaFileSchema
		}
	}

	return nil
}

func nestSchemaInScopes(schema *Schema, scopes []string) *Schema {
	scopeList := slices.Clone(scopes)
	slices.Reverse(scopeList)

	for _, scope := range scopeList {
		tmpSchema := &Schema{
			Properties: map[string]*Schema{
				scope: schema,
			},
		}
		schema = tmpSchema
	}
	return schema
}

// CreateJSONSchemaForChart is the public entry point for creating a JSON schema for a chart
func CreateJSONSchemaForChart(chart *charts.Chart, chartStore *charts.ChartStore, getSchemaPathForChart func(chart *charts.Chart) string) (GeneratedChartJSONSchema, error) {
	generator := NewSchemaGenerator(chart, chartStore, getSchemaPathForChart)
	return generator.Generate()
}
