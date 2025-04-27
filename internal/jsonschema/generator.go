package jsonschema

// generateJSONSchema generates a JSON schema from a map[string]interface{} instance
func generateJSONSchema(data map[string]any, description string) *Schema {
	schema := &Schema{
		Version:     Version,
		Type:        "object",
		Properties:  generateProperties(data, description),
		Description: description,
	}

	return schema
}

// generateProperties recursively inspects data and generates schema properties
func generateProperties(data map[string]interface{}, description string) map[string]*Schema {
	properties := make(map[string]*Schema)

	for key, value := range data {
		properties[key] = generateSchemaType(value, description)
		properties[key].Description = description
	}

	return properties
}

// generateSchemaType infers the JSON schema type based on the value's Go type
func generateSchemaType(value interface{}, description string) *Schema {
	schema := &Schema{}

	switch v := value.(type) {
	case string:
		schema.Type = "string"
		schema.Default = v

	case int, int32, int64:
		schema.Type = "integer"
		schema.Default = v
	case float32, float64:
		schema.Type = "number"
	case bool:
		schema.Type = "boolean"
		schema.Default = v
	case map[string]any:
		schema.Type = "object"
		schema.Properties = generateProperties(v, description)
	case []any:
		schema.Type = "array"
		if len(v) > 0 {
			// Only use the first element of the array to keep the schema simple.
			// Using allOf would create a possible large schema
			schema.Items = generateSchemaType(v[0], description)
		} else {
			schema.Items = &Schema{Type: "null"} // Default for empty arrays
		}
	default:
		schema.Type = "null"
	}

	return schema
}

func generateSchemaWithAllOf(definitions map[string]*Schema, references []*Schema) *Schema {
	schema := &Schema{
		Type:        "object",
		Version:     Version,
		Definitions: definitions,
		AllOf:       references,
	}
	return schema
}
