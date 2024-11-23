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

	schema, err := GenerateJSONSchema(input)
	assert.NoError(t, err)

	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	assert.NoError(t, err)

	// Expected JSON schema structure
	expected := map[string]interface{}{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type":    "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
			"age": map[string]interface{}{
				"type": "number",
			},
			"address": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"city": map[string]interface{}{"type": "string"},
					"zip":  map[string]interface{}{"type": "string"},
				},
			},
			"tags": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	expectedJSON, err := json.MarshalIndent(expected, "", "  ")
	assert.NoError(t, err)
	assert.Equal(t, string(expectedJSON), string(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to marshal expected schema to JSON: %v", err)
	}
}
