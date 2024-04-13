package util

import (
	"fmt"
	"reflect"

	"github.com/mrjosh/helm-ls/pkg/chartutil"
	"gopkg.in/yaml.v2"
)

func GetTableOrValueForSelector(values chartutil.Values, selector string) (string, error) {
	if len(selector) > 0 {
		localValues, err := values.Table(selector)
		if err != nil {
			value, err := values.PathValue(selector)
			return FormatToYAML(reflect.Indirect(reflect.ValueOf(value)), selector), err
		}
		return localValues.YAML()
	}
	return values.YAML()
}

func FormatToYAML(field reflect.Value, fieldName string) string {
	switch field.Kind() {
	case reflect.String:
		return field.String()
	case reflect.Map:
		return toYAML(field.Interface())
	case reflect.Slice:
		return toYAML(map[string]interface{}{fieldName: field.Interface()})
	case reflect.Bool:
		return fmt.Sprint(GetBoolType(field))
	case reflect.Float32, reflect.Float64:
		return fmt.Sprint(field.Float())
	default:
		return "<Unknown>"
	}
}

func toYAML(value interface{}) string {
	valBytes, _ := yaml.Marshal(value)
	return string(valBytes)
}

func GetBoolType(value interface{}) string {
	if val, ok := value.(bool); ok && val {
		return "True"
	}
	return "False"
}
