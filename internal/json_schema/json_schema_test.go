package jsonschema

import (
	"encoding/json"
	"testing"
)

func TestGenerateJSONSchema(t *testing.T) {
	// Define a sample input map to generate a schema from
	input := map[string]interface{}{
		"name":    "example",
		"age":     30,
		"address": map[string]interface{}{"city": "ExampleCity", "zip": "12345"},
		"tags":    []interface{}{"go", "json", "schema"},
	}

	// Generate the schema
	schema, err := GenerateJSONSchema(input)
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Marshal the schema to JSON
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal schema to JSON: %v", err)
	}

	// Print the JSON schema for visual verification
	t.Logf("Generated JSON Schema:\n%s", schemaJSON)

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

	// Convert the expected schema to JSON for comparison
	expectedJSON, err := json.MarshalIndent(expected, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal expected schema to JSON: %v", err)
	}

	// Check if generated schema matches the expected schema
	if string(schemaJSON) != string(expectedJSON) {
		t.Errorf("Generated schema does not match expected schema.\nExpected:\n%s\nGot:\n%s", expectedJSON, schemaJSON)
	}
}
