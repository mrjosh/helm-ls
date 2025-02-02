package jsonschema

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/log"
)

var logger = log.GetLogger()

func createJsonSchemaForChart(chart *charts.Chart, chartStore *charts.ChartStore) (string, error) {
	definitions := map[string]*Schema{}
	references := []*Schema{}
	globalDefs := []*Schema{}

	for _, scopedValuesfiles := range chart.GetScopedValuesFiles(chartStore) {
		innerSubSchemas := []*Schema{}
		for _, valuesFile := range scopedValuesfiles.ValuesFiles.AllValuesFiles() {

			subVals := valuesFile.Values.AsMap()
			for _, subScope := range scopedValuesfiles.SubScope {
				sub, ok := subVals[subScope].(map[string]interface{})
				if !ok || sub == nil {
					logger.Error("subscope value is nil", scopedValuesfiles.SubScope)
					subVals = map[string]interface{}{}
					continue
				}
				subVals = sub
			}

			scopeList := slices.Clone(scopedValuesfiles.Scope)
			slices.Reverse(scopeList)
			for _, scope := range scopeList {
				delete(subVals, "global") // IDEA: dont delete here, but instead to the global thing first
				tmpVals := make(map[string]interface{})
				tmpVals[scope] = subVals
				subVals = tmpVals
			}

			schema, err := generateJSONSchema(subVals)
			if err != nil {
				logger.Error(err)
				continue
			}
			if schema.Properties != nil && schema.Properties["global"] != nil {
				globalDefs = append(globalDefs, schema.Properties["global"])
				schema.Properties["global"] = &Schema{Ref: "#/$defs/global"}
			}

			if global := valuesFile.Values.AsMap()["global"]; global != nil {
				if reflect.TypeOf(global).Kind() == reflect.Map {
					globalCast := global.(map[string]interface{})
					globalSchema, err := generateJSONSchema(globalCast)
					if err != nil {
						logger.Error(err)
						continue
					}
					globalSchema.ID = scopedValuesfiles.Name
					globalDefs = append(globalDefs, globalSchema)
				}
			}

			innerSubSchemas = append(innerSubSchemas, schema)
		}

		if len(innerSubSchemas) > 0 {
			s := generateSchemaWithSubSchemas(innerSubSchemas)
			s.ID = scopedValuesfiles.Name
			definitions[scopedValuesfiles.Name] = s
		}

		references = append(references,
			&Schema{
				Ref: fmt.Sprintf("#/$defs/%s", scopedValuesfiles.Name),
			},
		)

	}

	subSchemas := []*Schema{}
	for _, v := range definitions {
		subSchemas = append(subSchemas, v)
	}

	references = append(references, &Schema{Ref: "#/$defs/global"})

	definitions["global"] = generateSchemaWithSubSchemas(globalDefs)

	schema := generateSchemaWithReferences(definitions, references)

	bytes, err := json.MarshalIndent(schema, "", "  ") // TODO: remove indent
	if err != nil {
		logger.Error(err)
		return "", err
	}

	file, err := os.CreateTemp("", base64.StdEncoding.EncodeToString([]byte(chart.RootURI.Filename()))+`*.json`)
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
