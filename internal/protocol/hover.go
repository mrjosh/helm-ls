package protocol

import (
	"fmt"
	"path/filepath"
	"sort"

	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

type HoverResultWithFile struct {
	Value string
	URI   uri.URI
}

type HoverResultsWithFiles []HoverResultWithFile

func (h HoverResultsWithFiles) FormatHelm(rootURI uri.URI) string {
	return h.format(rootURI, "helm")
}

func (h HoverResultsWithFiles) FormatYaml(rootURI uri.URI) string {
	return h.format(rootURI, "yaml")
}

func (h HoverResultsWithFiles) format(rootURI uri.URI, codeLanguage string) string {
	var formatted string
	sort.Slice(h, func(i, j int) bool {
		return h[i].URI > h[j].URI
	})

	for _, result := range h {
		value := result.Value
		if value == "" {
			value = "\"\""
		} else {
			value = fmt.Sprintf("```%s\n%s\n```", codeLanguage, value)
		}
		filepath, err := filepath.Rel(rootURI.Filename(), result.URI.Filename())
		if err != nil {
			filepath = result.URI.Filename()
		}
		formatted += fmt.Sprintf("### %s\n%s\n", filepath, value)
	}
	return formatted
}

func BuildHoverResponse(value string, wordRange lsp.Range) *lsp.Hover {
	content := lsp.MarkupContent{
		Kind:  lsp.Markdown,
		Value: value,
	}
	result := lsp.Hover{
		Contents: content,
		Range:    &wordRange,
	}
	return &result
}
