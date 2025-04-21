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
}

// NewSchemaGenerator creates a new SchemaGenerator instance
func NewSchemaGenerator(chart *charts.Chart, chartStore *charts.ChartStore, getSchemaPathForChart func(chart *charts.Chart) string) *SchemaGenerator {
	return &SchemaGenerator{
		chart:                 chart,
		chartStore:            chartStore,
		getSchemaPathForChart: getSchemaPathForChart,
	}
}

type GeneratedChartJSONSchema struct {
	schema       *Schema
	dependencies []*charts.Chart
}

func (g *SchemaGenerator) Generate() (GeneratedChartJSONSchema, error) {
	dependencies := []*charts.Chart{}
	definitions := map[string]*Schema{}
	globalSchemas := []*Schema{}
	allOf := []*Schema{}

	// Process all scoped values files
	for _, scopedValuesfiles := range g.chart.GetScopedValuesFiles(g.chartStore) {
		if len(scopedValuesfiles.Scope) == 0 && len(scopedValuesfiles.SubScope) == 0 {
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
							globalSchemas = append(globalSchemas, globalSchema)
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
			if scopedValuesfiles.Chart.HelmChart.Schema != nil {

				schemaFileSchema := &Schema{}
				err := json.Unmarshal(scopedValuesfiles.Chart.HelmChart.Schema, schemaFileSchema)
				if err != nil {
					logger.Error("Failed to unmarshal schema from helm chart "+scopedValuesfiles.Chart.RootURI, err)
				} else {
					valuesSchemas = append(valuesSchemas, schemaFileSchema)
				}
			}
			definitions[scopedValuesfiles.Chart.HelmChart.Name()] = &Schema{
				AllOf: valuesSchemas,
			}
			allOf = append(allOf, &Schema{
				Ref: fmt.Sprintf("#/$defs/%s", scopedValuesfiles.Chart.HelmChart.Name()),
			})
			allOf = append(allOf, &Schema{
				Ref: "#/$defs/global",
			})
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
					_, err := g.getSubScope(vals, scopedValuesfiles.SubScope)
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
				schema = g.nestSchemaInScopes(schema, scopedValuesfiles.Scope)
				allOf = append(allOf, schema)
			}
			globalSchemaRef := fmt.Sprintf("%s#/$defs/%s",
				schemFilePath,
				"global",
			)
			globalSchema := &Schema{Ref: globalSchemaRef}
			allOf = append(allOf, g.nestSchemaInScopes(globalSchema, []string{"global"}))
		}
	}
	definitions["global"] = &Schema{
		AllOf: globalSchemas,
	}

	schema := generateSchemaWithAllOf(definitions, allOf)

	return GeneratedChartJSONSchema{
		schema:       schema,
		dependencies: dependencies,
	}, nil // TODO: collect errors
}

// getSubScope returns the values for the given subscope
// e.g. given values: {a: {b: {c: 1}}}, subScopes: ["a", "b"] returns {c: 1}
func (g *SchemaGenerator) getSubScope(values map[string]any, subScopes []string) (map[string]interface{}, error) {
	for _, subScope := range subScopes {
		sub, ok := values[subScope].(map[string]interface{})
		if !ok || sub == nil {
			return nil, fmt.Errorf("subscope value is nil for scope: %s", subScope)
		}
		values = sub
	}
	return values, nil
}

func (g *SchemaGenerator) nestSchemaInScopes(schema *Schema, scopes []string) *Schema {
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
