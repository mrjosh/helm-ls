package jsonschema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateJSONSchema(t *testing.T) {
	input := map[string]any{
		"name":    "example",
		"age":     30,
		"address": map[string]any{"city": "ExampleCity", "zip": "12345"},
		"tags":    []any{"go", "json", "schema"},
	}

	schema := generateJSONSchema(input, "description")

	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	assert.NoError(t, err)

	expected := &Schema{
		Version: "https://json-schema.org/draft/2020-12/schema",
		Type:    Type{"object"},
		Properties: map[string]*Schema{
			"name": {Type: Type{"string"}, Default: "example", Description: "description"},
			"age":  {Type: Type{"integer"}, Default: 30, Description: "description"},
			"address": {
				Type: Type{"object"},
				Properties: map[string]*Schema{
					"city": {
						Type:        Type{"string"},
						Description: "description",
						Default:     "ExampleCity",
					},
					"zip": {
						Type: Type{"string"}, Default: "12345",
						Description: "description",
					},
				},
				Description: "description",
			},
			"tags": {Type: Type{"array"}, Items: &Schema{Type: Type{"string"}, Default: "go"}, Description: "description"},
		},
		Description: "description",
	}

	expectedJSON, err := json.MarshalIndent(expected, "", "  ")

	assert.NoError(t, err)
	assert.Equal(t, string(expectedJSON), string(schemaJSON))
	if err != nil {
		t.Fatalf("Failed to marshal expected schema to JSON: %v", err)
	}
}
