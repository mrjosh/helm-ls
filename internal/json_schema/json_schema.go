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
	// reflector := jsonschema.Reflector{
	// 	ExpandedStruct:            true,
	// 	AllowAdditionalProperties: true,
	// }

	schema, err := GenerateJSONSchema(chart.ValuesFiles.MainValuesFile.Values)

	bytes, err := json.Marshal(schema)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	// create a tmp file and write the schema

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

// func GenerateSchemaFromData(data interface{}) error {
// 	jsonBytes, err := json.Marshal(data)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	documentLoader := gojsonschema.NewStringLoader(string(jsonBytes))
// 	schema, err := gojsonschema.NewSchema(documentLoader)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return schema.Root(), nil
// }
