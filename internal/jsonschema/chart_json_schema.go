package jsonschema

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/log"
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
			schemFilePath := uri.File(g.getSchemaPathForChart(scopedValuesfiles.Chart))
			dependencies = append(dependencies, scopedValuesfiles.Chart)

			refPointers := []string{""}

			if len(scopedValuesfiles.SubScope) > 0 {
				refPointers = []string{}
				refPointer := ""
				for _, v := range scopedValuesfiles.SubScope {
					refPointer = fmt.Sprintf("%s/properties/%s", refPointer, v)
				}

				for i, valuesFile := range scopedValuesfiles.ValuesFiles.AllValuesFiles() {
					vals := valuesFile.Values.AsMap()
					_, err := getSubScope(vals, scopedValuesfiles.SubScope)
					if err == nil {
						refPointers = append(refPointers, fmt.Sprintf("/allOf/%d%s", i, refPointer))
					}
				}
			}

			for _, refPointer := range refPointers {
				ref := fmt.Sprintf("%s#/$defs/%s%s",
					schemFilePath,
					scopedValuesfiles.Chart.HelmChart.Name(),
					refPointer,
				)

				schema := &Schema{Ref: ref}
				schema = nestSchemaInScopes(schema, scopedValuesfiles.Scope)
				g.allOf = append(g.allOf, schema)
			}
			globalSchemaRef := fmt.Sprintf("%s#/$defs/%s",
				schemFilePath,
				"global",
			)
			globalSchema := &Schema{Ref: globalSchemaRef}
			g.allOf = append(g.allOf, nestSchemaInScopes(globalSchema, []string{"global"}))
		}
	}

	g.addGlobalDef()

	schema := generateSchemaWithAllOf(g.definitions, g.allOf)

	return GeneratedChartJSONSchema{
		schema:       schema,
		dependencies: dependencies,
	}, nil // TODO: collect errors
}

func (g *SchemaGenerator) addGlobalDef() {
	g.definitions["global"] = &Schema{
		AllOf: g.globalSchemas,
	}
	g.allOf = append(g.allOf, &Schema{
		Ref: "#/$defs/global",
	})
}

func (g *SchemaGenerator) generateSchemaForCurrentChart(scopedValuesfiles *charts.ScopedValuesFiles) {
	valuesSchemas := []*Schema{}
	for _, valuesFile := range scopedValuesfiles.ValuesFiles.AllValuesFiles() {
		subVals := valuesFile.Values.AsMap()

		globalVals, ok := subVals["global"]
		if ok {
			subValsTmp := map[string]any{}
			for k, v := range subVals {
				if k != "global" {
					subValsTmp[k] = v
				}
			}
			subVals = subValsTmp

			globalValsMap, ok := globalVals.(map[string]any)
			if ok {
				globalSchema, err := generateJSONSchema(globalValsMap, "global values from the file "+filepath.Base(valuesFile.URI.Filename()))
				if err != nil {
					logger.Error("Failed to generate JSON schema:", err)
				} else {
					g.globalSchemas = append(g.globalSchemas, globalSchema)
				}
			}
		}

		schema, err := generateJSONSchema(subVals,
			fmt.Sprintf("%s values from the file %s",
				scopedValuesfiles.Chart.HelmChart.Name(),
				filepath.Base(valuesFile.URI.Filename())))
		if err != nil {
			logger.Error("Failed to generate JSON schema:", err)
			continue
		}

		valuesSchemas = append(valuesSchemas, schema)
	}

	if schemaFileSchema := getSchemaFileSchema(scopedValuesfiles.Chart); schemaFileSchema != nil {
		valuesSchemas = append(valuesSchemas, schemaFileSchema)
	}

	g.definitions[scopedValuesfiles.Chart.HelmChart.Name()] = &Schema{
		AllOf: valuesSchemas,
	}
	g.allOf = append(g.allOf, &Schema{
		Ref: fmt.Sprintf("#/$defs/%s", scopedValuesfiles.Chart.HelmChart.Name()),
	})
}

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

// getSubScope returns the values for the given subscope
// e.g. given values: {a: {b: {c: 1}}}, subScopes: ["a", "b"] returns {c: 1}
func getSubScope(values map[string]any, subScopes []string) (map[string]any, error) {
	for _, subScope := range subScopes {
		sub, ok := values[subScope].(map[string]any)
		if !ok || sub == nil {
			return nil, fmt.Errorf("subscope value is nil for scope: %s", subScope)
		}
		values = sub
	}
	return values, nil
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
