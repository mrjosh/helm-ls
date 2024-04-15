package util

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mrjosh/helm-ls/pkg/chartutil"
	lsp "go.lsp.dev/protocol"
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

func GetValueCompletion(values chartutil.Values, splittedVar []string) []lsp.CompletionItem {
	var (
		err         error
		tableName   = strings.Join(splittedVar, ".")
		localValues chartutil.Values
		items       = make([]lsp.CompletionItem, 0)
	)

	if len(splittedVar) > 0 {

		localValues, err = values.Table(tableName)
		if err != nil {
			if len(splittedVar) > 1 {
				// the current tableName was not found, maybe because it is incomplete, we can use the previous one
				// e.g. gobal.im -> im was not found
				// but global contains the key image, so we return all keys of global
				localValues, err = values.Table(strings.Join(splittedVar[:len(splittedVar)-1], "."))
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
		items = setItem(items, value, variable)
	}

	return items
}

func setItem(items []lsp.CompletionItem, value interface{}, variable string) []lsp.CompletionItem {
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

	return append(items, lsp.CompletionItem{
		Label:         variable,
		InsertText:    variable,
		Documentation: documentation,
		Detail:        valueOf.Kind().String(),
		Kind:          itemKind,
	})
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
