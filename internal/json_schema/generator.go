package jsonschema

// generateJSONSchema generates a JSON schema from a map[string]interface{} instance
func generateJSONSchema(data map[string]interface{}) (*Schema, error) {
	schema := &Schema{
		Version:    Version,
		Type:       "object",
		Properties: generateProperties(data),
	}

	return schema, nil
}

// generateProperties recursively inspects data and generates schema properties
func generateProperties(data map[string]interface{}) map[string]*Schema {
	properties := make(map[string]*Schema)

	for key, value := range data {
		properties[key] = generateSchemaType(value)
	}

	return properties
}

// generateSchemaType infers the JSON schema type based on the value's Go type
func generateSchemaType(value interface{}) *Schema {
	schema := &Schema{}

	switch v := value.(type) {
	case string:
		schema.Type = "string"
		schema.Default = v

	case int, int32, int64, float32, float64:
		schema.Type = "number"
		schema.Default = v
	case bool:
		schema.Type = "boolean"
		schema.Default = v
	case map[string]interface{}:
		schema.Type = "object"
		schema.Properties = generateProperties(v)
	case []interface{}:
		schema.Type = "array"
		if len(v) > 0 {
			schema.Items = generateSchemaType(v[0])
		} else {
			schema.Items = &Schema{Type: "null"} // Default for empty arrays
		}
	default:
		// schema["type"] = "null" // Fallback to null for unknown types
		schema.Type = "null"
	}

	return schema
}

func generateSchemaWithSubSchemas(subSchemas []*Schema) *Schema {
	schema := &Schema{
		Type:    "object",
		Version: Version,
		AllOf:   subSchemas,
	}
	return schema
}
