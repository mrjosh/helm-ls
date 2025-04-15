package jsonschema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateJSONSchema(t *testing.T) {
	input := map[string]interface{}{
		"name":    "example",
		"age":     30,
		"address": map[string]interface{}{"city": "ExampleCity", "zip": "12345"},
		"tags":    []interface{}{"go", "json", "schema"},
	}

	schema, err := generateJSONSchema(input, "test schema")
	assert.NoError(t, err)

	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	assert.NoError(t, err)

	expected := &Schema{
		Version: "https://json-schema.org/draft/2020-12/schema",
		Type:    "object",
		Properties: map[string]*Schema{
			"name": {Type: "string", Default: "example"},
			"age":  {Type: "number", Default: 30},
			"address": {
				Type: "object",
				Properties: map[string]*Schema{
					"city": {
						Type: "string", Default: "ExampleCity",
					},
					"zip": {
						Type: "string", Default: "12345",
					},
				},
			},
			"tags": {Type: "array", Items: &Schema{Type: "string", Default: "go"}},
		},
	}

	expectedJSON, err := json.MarshalIndent(expected, "", "  ")

	assert.NoError(t, err)
	assert.Equal(t, string(expectedJSON), string(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to marshal expected schema to JSON: %v", err)
	}
}
