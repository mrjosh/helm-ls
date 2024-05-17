package protocol

import (
	"github.com/mrjosh/helm-ls/internal/documentation/godocs"
	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
	lsp "go.lsp.dev/protocol"
)

type CompletionResults struct {
	Items []lsp.CompletionItem
}

func (c CompletionResults) ToList() (result *lsp.CompletionList) {
	return &lsp.CompletionList{Items: c.Items, IsIncomplete: false}
}

func (c CompletionResults) WithDocs(docs []helmdocs.HelmDocumentation, kind lsp.CompletionItemKind) CompletionResults {
	items := c.Items
	for _, doc := range docs {
		items = append(items,
			lsp.CompletionItem{
				Label:            doc.Name,
				Detail:           doc.Detail,
				InsertText:       doc.Name,
				InsertTextFormat: lsp.InsertTextFormatPlainText,
				Kind:             kind,
			},
		)
	}
	return CompletionResults{Items: items}
}

func (c CompletionResults) WithSnippets(snippets []godocs.GoTemplateSnippet) CompletionResults {
	items := c.Items
	for _, snippet := range snippets {
		items = append(items, snippetCompletionItem(snippet))
	}
	return CompletionResults{Items: items}
}

func snippetCompletionItem(gotemplateSnippet godocs.GoTemplateSnippet) lsp.CompletionItem {
	return lsp.CompletionItem{
		Label:            gotemplateSnippet.Name,
		InsertText:       gotemplateSnippet.Snippet,
		Detail:           gotemplateSnippet.Detail,
		Documentation:    gotemplateSnippet.Doc,
		Kind:             lsp.CompletionItemKindText,
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	}
}
