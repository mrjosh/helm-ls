package protocol

import (
	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
	lsp "go.lsp.dev/protocol"
)

type CompletionResults []CompletionResult

func NewCompletionResults(docs []helmdocs.HelmDocumentation) *CompletionResults {
	result := CompletionResults{}

	for _, doc := range docs {
		result = append(result, CompletionResult{doc})
	}

	return &result
}

type CompletionResult struct {
	Documentation helmdocs.HelmDocumentation
}

func (c *CompletionResult) ToLSP() (result lsp.CompletionItem) {
	return lsp.CompletionItem{
		Label:            c.Documentation.Name,
		Detail:           c.Documentation.Detail,
		InsertText:       c.Documentation.Name,
		InsertTextFormat: lsp.InsertTextFormatSnippet,
		Kind:             lsp.CompletionItemKindConstant, // TODO: make this more variable
	}
}

func (c *CompletionResults) ToLSP() (result *lsp.CompletionList) {
	items := []lsp.CompletionItem{}

	for _, completion := range *c {
		items = append(items, completion.ToLSP())
	}

	return &lsp.CompletionList{Items: items}
}
