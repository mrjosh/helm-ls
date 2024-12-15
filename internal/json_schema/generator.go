package jsonschema

// GenerateJSONSchema generates a JSON schema from a map[string]interface{} instance
func GenerateJSONSchema(data map[string]interface{}) (map[string]interface{}, error) {
	schema := map[string]interface{}{
		"$schema":    "https://json-schema.org/draft/2020-12/schema",
		"type":       "object",
		"properties": generateProperties(data),
	}

	return schema, nil
}

// generateProperties recursively inspects data and generates schema properties
func generateProperties(data map[string]interface{}) map[string]interface{} {
	properties := make(map[string]interface{})

	for key, value := range data {
		properties[key] = generateSchemaType(value)
	}

	return properties
}

// generateSchemaType infers the JSON schema type based on the value's Go type
func generateSchemaType(value interface{}) map[string]interface{} {
	schema := make(map[string]interface{})

	switch v := value.(type) {
	case string:
		schema["type"] = "string"
	case int, int32, int64, float32, float64:
		schema["type"] = "number"
	case bool:
		schema["type"] = "boolean"
	case map[string]interface{}:
		schema["type"] = "object"
		schema["properties"] = generateProperties(v)
	case []interface{}:
		schema["type"] = "array"
		if len(v) > 0 {
			schema["items"] = generateSchemaType(v[0])
		} else {
			schema["items"] = map[string]interface{}{"type": "null"} // Default for empty arrays
		}
	default:
		schema["type"] = "null" // Fallback to null for unknown types
	}

	return schema
}
