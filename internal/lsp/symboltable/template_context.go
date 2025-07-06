package symboltable

import (
	"regexp"
	"strings"
)

type TemplateContext []string

func (t TemplateContext) Format() string {
	return strings.Join(t, ".")
}

func (t TemplateContext) Copy() TemplateContext {
	return append(TemplateContext{}, t...)
}

// Return everything except the first context
func (t TemplateContext) Tail() TemplateContext {
	if len(t) == 0 {
		return t
	}
	return t[1:]
}

func (t TemplateContext) IsVariable() bool {
	return len(t) > 0 && strings.HasPrefix(t[0], "$")
}

// Adds a suffix to the last context
func (t TemplateContext) AppendSuffix(suffix string) TemplateContext {
	if len(t) == 0 {
		return TemplateContext{suffix}
	}
	t[len(t)-1] = t[len(t)-1] + suffix
	return t
}

func NewTemplateContext(string string) TemplateContext {
	if string == "." {
		return TemplateContext{}
	}
	splitted := strings.Split(string, ".")
	if len(splitted) > 0 && splitted[0] == "" {
		return splitted[1:]
	}
	return splitted
}

// Converts a YAML Path from the  goccy go-yaml library to a template context
// From: https://github.com/goccy/go-yaml/blob/v1.18.0/path.go#L27
// // YAMLPath rule
// $     : the root object/element
// .     : child operator
// ..    : recursive descent (not supported)
// [num] : object/element of array by number
// [*]   : all objects/elements for array. (not supported)
func TemplateContextFromYAMLPath(jsonPath string) TemplateContext {
	spliited := strings.Split(jsonPath, ".")

	if len(spliited) > 0 && spliited[0] == "$" {
		spliited = spliited[1:]
	}
	for i := range spliited {
		spliited[i] = convertYAMLPathElement(spliited[i])
	}
	return spliited
}

var yamlPathIndexRegex = regexp.MustCompile(`\[\d+\]$`)

func convertYAMLPathElement(element string) string {
	if !strings.HasSuffix(element, "]") {
		return element
	}
	return yamlPathIndexRegex.ReplaceAllString(element, "[]")
}
