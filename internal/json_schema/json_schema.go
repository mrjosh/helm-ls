package jsonschema

import (
	"encoding/base64"
	"encoding/json"
	"os"

	// "github.com/invopop/jsonschema"
	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/log"
)

var logger = log.GetLogger()

func CreateJsonSchemaForChart(chart *charts.Chart) (string, error) {
	schema, err := GenerateJSONSchema(chart.ValuesFiles.MainValuesFile.Values)

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
