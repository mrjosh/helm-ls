package util

import (
	"fmt"
	"reflect"
	"strings"

	lsp "go.lsp.dev/protocol"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chartutil"
)

func GetTableOrValueForSelector(values chartutil.Values, selector []string) (string, error) {
	if len(selector) <= 0 || selector[0] == "" {
		return values.YAML()
	}
	value, err := pathLookup(values, selector)
	return FormatToYAML(reflect.Indirect(reflect.ValueOf(value)), strings.Join(selector, ".")), err
}

func GetValueCompletion(values chartutil.Values, splittedVar []string) []lsp.CompletionItem {
	var (
		err         error
		localValues chartutil.Values
		items       = make([]lsp.CompletionItem, 0)
	)

	if len(splittedVar) > 0 {
		localValues, err = valuesLookup(values, splittedVar)
		if err != nil {
			if len(splittedVar) > 1 {
				// the current tableName was not found, maybe because it is incomplete, we can use the previous one
				// e.g. gobal.im -> im was not found
				// but global contains the key image, so we return all keys of global
				localValues, err = valuesLookup(values, splittedVar[:len(splittedVar)-1])
				if err != nil {
					return []lsp.CompletionItem{}
				}
				values = localValues
			}
		} else {
			values = localValues
		}
	}

	for variable, value := range values {
		items = append(items, builCompletionItem(value, variable))
	}

	return items
}

func valuesLookup(values chartutil.Values, splittedVar []string) (chartutil.Values, error) {
	result, err := pathLookup(values, splittedVar)
	if err != nil {
		return chartutil.Values{}, err
	}

	values, ok := result.(map[string]interface{})
	if ok {
		return values, nil
	}

	return chartutil.Values{}, chartutil.ErrNoTable{Key: splittedVar[0]}
}

// PathValue takes a path that traverses a YAML structure and returns the value at the end of that path.
// The path starts at the root of the YAML structure and is comprised of YAML keys separated by periods.
// Given the following YAML data the value at path "chapter.one.title" is "Loomings". The path can also
// include array indexes as in "chapters[].title" which will use the first element of the array.
//
//	chapter:
//	  one:
//	    title: "Loomings"
func pathLookup(v chartutil.Values, path []string) (interface{}, error) {
	if len(path) == 0 {
		return v, nil
	}
	if strings.HasSuffix(path[0], "[]") {
		return arrayLookup(v, path)
	}
	// if exists must be root key not table
	value, ok := v[path[0]]
	if !ok {
		return nil, chartutil.ErrNoTable{Key: path[0]}
	}
	if len(path) == 1 {
		return value, nil
	}
	if nestedValues, ok := value.(map[string]interface{}); ok {
		return pathLookup(nestedValues, path[1:])
	}
	return nil, chartutil.ErrNoTable{Key: path[0]}
}

func arrayLookup(v chartutil.Values, path []string) (interface{}, error) {
	v2, ok := v[path[0][:(len(path[0])-2)]]
	if !ok {
		return v, chartutil.ErrNoTable{Key: fmt.Sprintf("Yaml key %s does not exist", path[0])}
	}
	if v3, ok := v2.([]interface{}); ok {
		if len(v3) == 0 {
			return chartutil.Values{}, ErrEmpytArray{path[0]}
		}
		if len(path) == 1 {
			return v3[0], nil
		}
		if vv, ok := v3[0].(map[string]interface{}); ok {
			return pathLookup(vv, path[1:])
		}
		return chartutil.Values{}, chartutil.ErrNoTable{Key: path[0]}
	}
	if nestedValues, ok := v2.(map[string]interface{}); ok {
		if len(nestedValues) == 0 {
			return chartutil.Values{}, ErrEmpytMapping{path[0]}
		}

		for k := range nestedValues {
			if len(path) == 1 {
				return nestedValues[k], nil
			}
			if nestedValues, ok := (nestedValues[k]).(map[string]interface{}); ok {
				return pathLookup(nestedValues, path[1:])
			}
		}
		return chartutil.Values{}, chartutil.ErrNoTable{Key: path[0]}
	}

	return chartutil.Values{}, chartutil.ErrNoTable{Key: path[0]}
}

type ErrEmpytArray struct {
	Key string
}
type ErrEmpytMapping struct {
	Key string
}

func (e ErrEmpytArray) Error() string   { return fmt.Sprintf("%q is an empyt array", e.Key) }
func (e ErrEmpytMapping) Error() string { return fmt.Sprintf("%q is an empyt mapping", e.Key) }

func builCompletionItem(value interface{}, variable string) lsp.CompletionItem {
	var (
		itemKind      = lsp.CompletionItemKindVariable
		valueOf       = reflect.ValueOf(value)
		documentation = valueOf.String()
	)

	switch valueOf.Kind() {
	case reflect.Slice, reflect.Map:
		itemKind = lsp.CompletionItemKindStruct
		documentation = toYAML(value)
	case reflect.Bool:
		itemKind = lsp.CompletionItemKindVariable
		documentation = GetBoolType(value)
	case reflect.Float32, reflect.Float64:
		documentation = fmt.Sprintf("%.2f", valueOf.Float())
		itemKind = lsp.CompletionItemKindVariable
	case reflect.Invalid:
		documentation = "<Unknown>"
	default:
		itemKind = lsp.CompletionItemKindField
	}

	return lsp.CompletionItem{
		Label:         variable,
		InsertText:    variable,
		Documentation: documentation,
		Detail:        valueOf.Kind().String(),
		Kind:          itemKind,
	}
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
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprint(field.Int())
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
