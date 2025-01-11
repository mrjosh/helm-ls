package jsonschema

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/log"
)

var logger = log.GetLogger()

func createJsonSchemaForChart(chart *charts.Chart) (string, error) {
	subSchemas := []*Schema{}
	for _, value := range chart.ValuesFiles.AdditionalValuesFiles {
		if value == nil || len(value.Values) == 0 {
			continue
		}
		subSchema, err := generateJSONSchema(value.Values)
		if err != nil {
			logger.Error(err)
			continue
		}

		subSchemas = append(subSchemas, subSchema)
	}

	if len(subSchemas) == 0 {
		return "", errors.New("No values found to generate schema for")
	}

	schema := generateSchemaWithSubSchemas(subSchemas)

	bytes, err := json.Marshal(schema)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	file, err := os.CreateTemp("", base64.StdEncoding.EncodeToString([]byte(chart.RootURI.Filename())))
	if err != nil {
		logger.Error(err)
		return "", err
	}

	_, err = file.Write(bytes)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	return file.Name(), nil
}
