package util

import lsp "go.lsp.dev/protocol"

func BuildHoverResponse(value string, wordRange lsp.Range) lsp.Hover {
	content := lsp.MarkupContent{
		Kind:  lsp.Markdown,
		Value: value,
	}
	result := lsp.Hover{
		Contents: content,
		Range:    &wordRange,
	}
	return result
}
