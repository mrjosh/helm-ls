package protocol

import (
	"github.com/mrjosh/helm-ls/internal/documentation/godocs"
	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
	"github.com/mrjosh/helm-ls/internal/lsp/symboltable"
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
				Detail:           doc.Detail + "\n\n" + doc.Doc,
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

func (c CompletionResults) WithVariableDefinitions(variableDefinitions map[string][]symboltable.VariableDefinition) CompletionResults {
	items := c.Items
	for variableName, definitions := range variableDefinitions {

		if len(definitions) == 0 {
			continue
		}
		definition := definitions[0]

		items = append(items,
			variableCompletionItem(variableName, definition),
		)
	}
	return CompletionResults{Items: items}
}

func variableCompletionItem(variableName string, definition symboltable.VariableDefinition) lsp.CompletionItem {
	return lsp.CompletionItem{
		Label:            variableName,
		Detail:           definition.Value,
		InsertText:       variableName,
		InsertTextFormat: lsp.InsertTextFormatPlainText,
		Kind:             lsp.CompletionItemKindVariable,
	}
}
