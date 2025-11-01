package jsonschema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateJSONSchemaTestdata(t *testing.T) {
	count := 0
	err := filepath.Walk("../../testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == "values.schema.json" {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var input map[string]any
			if err = json.Unmarshal(data, &input); err != nil {
				return err
			}

			schemaFileSchema := &Schema{}
			unmarshalErr := json.Unmarshal(data, schemaFileSchema)

			assert.NoError(t, unmarshalErr)
			count++
		}

		return nil
	})
	assert.True(t, count == 1 || count == 23) // one if the testdata does not include bitnami charts/ 23 if it does
	assert.NoError(t, err)
}
