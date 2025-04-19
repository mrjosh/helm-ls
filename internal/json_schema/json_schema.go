package jsonschema

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/log"
)

var logger = log.GetLogger()

// SchemaGenerator handles the generation of JSON schemas for Helm charts
type SchemaGenerator struct {
	chart      *charts.Chart
	chartStore *charts.ChartStore
}

// NewSchemaGenerator creates a new SchemaGenerator instance
func NewSchemaGenerator(chart *charts.Chart, chartStore *charts.ChartStore) *SchemaGenerator {
	return &SchemaGenerator{
		chart:      chart,
		chartStore: chartStore,
	}
}

// Generate creates a JSON schema file for the chart and returns its path
func (g *SchemaGenerator) Generate() (string, error) {
	definitions := make(map[string]*Schema)
	references := []*Schema{}
	globalDefs := []*Schema{}

	// Process all scoped values files
	for _, scopedValuesfiles := range g.chart.GetScopedValuesFiles(g.chartStore) {
		innerSubSchemas, err := g.processScopedValuesFiles(scopedValuesfiles, &globalDefs)
		if err != nil {
			logger.Error("Failed to process scoped values files:", err)
			continue
		}

		if len(innerSubSchemas) > 0 {
			schema := g.createDefinitionSchema(scopedValuesfiles.Name, innerSubSchemas)
			definitions[scopedValuesfiles.Name] = schema
			references = append(references, g.nestSchemaInScopes(&Schema{
				Ref: fmt.Sprintf("#/$defs/%s", scopedValuesfiles.Name),
			}, scopedValuesfiles.Scope))
		}
	}

	// Add global definitions
	references = append(references, &Schema{Ref: "#/$defs/global"})
	definitions["global"] = g.createGlobalSchema(globalDefs)

	// Generate final schema
	schema := generateSchemaWithReferences(definitions, references)
	return g.writeSchemaToFile(schema)
}

// processScopedValuesFiles processes a single set of scoped values files
func (g *SchemaGenerator) processScopedValuesFiles(scopedValuesfiles *charts.ScopedValuesFiles, globalDefs *[]*Schema) ([]*Schema, error) {
	innerSubSchemas := []*Schema{}

	for _, valuesFile := range scopedValuesfiles.ValuesFiles.AllValuesFiles() {
		subVals := valuesFile.Values.AsMap()

		// Handle global values
		if err := g.processGlobalValues(subVals, globalDefs); err != nil {
			logger.Error("Failed to process global values:", err)
		}

		// Process subscopes
		subVals, err := g.getSubScope(subVals, scopedValuesfiles.SubScope)
		if err != nil {
			logger.Error("Failed to process subscopes:", err)
			continue
		}

		// Generate schema for processed values
		schema, err := generateJSONSchema(subVals, fmt.Sprintf("%s values from the file %s", scopedValuesfiles.Name, filepath.Base(valuesFile.URI.Filename())))
		if err != nil {
			return nil, fmt.Errorf("failed to generate JSON schema: %w", err)
		}
		innerSubSchemas = append(innerSubSchemas, schema)
	}

	if scopedValuesfiles.Schema != nil {
		// parse the json schema
		schema := &Schema{}
		err := json.Unmarshal(scopedValuesfiles.Schema, schema)
		if err != nil {
			return innerSubSchemas, fmt.Errorf("failed to unmarshal schema: %w", err)
		}

		innerSubSchemas = append(innerSubSchemas, schema)
	}

	return innerSubSchemas, nil
}

// processGlobalValues extracts and processes global values from the given values map
func (g *SchemaGenerator) processGlobalValues(values map[string]interface{}, globalDefs *[]*Schema) error {
	globalVals, ok := values["global"]
	if !ok {
		return nil
	}

	globalValsMap, ok := globalVals.(map[string]interface{})
	if !ok {
		return fmt.Errorf("global values is not a map")
	}

	delete(values, "global")
	globalSchema, err := generateJSONSchema(globalValsMap, "global values")
	if err != nil {
		return fmt.Errorf("failed to generate global schema: %w", err)
	}

	*globalDefs = append(*globalDefs, globalSchema)
	return nil
}

// getSubScope returns the values for the given subscope
// e.g. given values: {a: {b: {c: 1}}}, subScopes: ["a", "b"] returns {c: 1}
func (g *SchemaGenerator) getSubScope(values map[string]interface{}, subScopes []string) (map[string]interface{}, error) {
	for _, subScope := range subScopes {
		sub, ok := values[subScope].(map[string]interface{})
		if !ok || sub == nil {
			return nil, fmt.Errorf("subscope value is nil for scope: %s", subScope)
		}
		values = sub
	}
	return values, nil
}

// nestValuesInScopes nests the given values by the given scopes
// e.g. given values: {a: 1}, scopes: ["b", "c"] returns {b: {c: {a: 1}}}
func (g *SchemaGenerator) nestValuesInScopes(values map[string]interface{}, scopes []string) map[string]interface{} {
	scopeList := slices.Clone(scopes)
	slices.Reverse(scopeList)

	for _, scope := range scopeList {
		tmpVals := make(map[string]interface{})
		tmpVals[scope] = values
		values = tmpVals
	}
	return values
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

// createDefinitionSchema creates a schema definition with the given name and subschemas
func (g *SchemaGenerator) createDefinitionSchema(name string, subSchemas []*Schema) *Schema {
	schema := generateSchemaWithSubSchemas(subSchemas)
	schema.ID = name
	return schema
}

// createGlobalSchema creates the global schema definition
func (g *SchemaGenerator) createGlobalSchema(globalDefs []*Schema) *Schema {
	return &Schema{
		Properties: map[string]*Schema{
			"global": generateSchemaWithSubSchemas(globalDefs),
		},
	}
}

// writeSchemaToFile writes the schema to a temporary file and returns its path
func (g *SchemaGenerator) writeSchemaToFile(schema *Schema) (string, error) {
	bytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}

	file, err := os.CreateTemp("", base64.StdEncoding.EncodeToString([]byte(g.chart.RootURI.Filename()))+`*.json`)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(bytes); err != nil {
		return "", fmt.Errorf("failed to write schema to file: %w", err)
	}

	return file.Name(), nil
}

// CreateJsonSchemaForChart is the public entry point for creating a JSON schema for a chart
func CreateJsonSchemaForChart(chart *charts.Chart, chartStore *charts.ChartStore) (string, error) {
	generator := NewSchemaGenerator(chart, chartStore)
	return generator.Generate()
}
