package languagefeatures

import (
	"fmt"
	"strings"

	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
	"github.com/mrjosh/helm-ls/internal/lsp/symboltable"
	"github.com/mrjosh/helm-ls/internal/protocol"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

func (f *TemplateContextFeature) Completion() (*lsp.CompletionList, error) {
	templateContext, err := f.getTemplateContext()
	if err != nil {
		return nil, err
	}

	if len(templateContext) == 0 {
		return protocol.CompletionResults{}.WithDocs(helmdocs.BuiltInObjects, lsp.CompletionItemKindConstant).ToList(), nil
	}

	if len(templateContext) == 1 {
		result, ok := helmdocs.BuiltInOjectVals[templateContext[0]]
		if !ok {
			return protocol.CompletionResults{}.WithDocs(helmdocs.BuiltInObjects, lsp.CompletionItemKindConstant).ToList(), nil
		}
		return protocol.CompletionResults{}.WithDocs(result, lsp.CompletionItemKindValue).ToList(), nil
	}
	if templateContext[0] == "Values" {
		return f.valuesCompletion(templateContext)
	}
	nestedDocs, ok := helmdocs.BuiltInOjectVals[templateContext[0]]
	if ok {
		if len(templateContext) < 3 {
			return protocol.CompletionResults{}.WithDocs(nestedDocs, lsp.CompletionItemKindValue).ToList(), nil
		}

		adjustedDocs := trimPrefixForNestedDocs(nestedDocs, templateContext)
		return protocol.CompletionResults{}.WithDocs(adjustedDocs, lsp.CompletionItemKindValue).ToList(), nil
	}
	return protocol.CompletionResults{}.ToList(), fmt.Errorf("%s is no valid template context for helm", templateContext)
}

// handels the completion for .Capabilities.KubeVersion.^ where the result should not contain KubeVersion again
func trimPrefixForNestedDocs(nestedDocs []helmdocs.HelmDocumentation, templateContext symboltable.TemplateContext) []helmdocs.HelmDocumentation {
	adjustedDocs := []helmdocs.HelmDocumentation{}
	for _, v := range nestedDocs {
		if strings.HasPrefix(v.Name, templateContext.Tail().Format()) {
			v.Name = strings.TrimPrefix(v.Name, templateContext.Tail().Format())
			adjustedDocs = append(adjustedDocs, v)
		}
	}
	return adjustedDocs
}

func (f *TemplateContextFeature) valuesCompletion(templateContext symboltable.TemplateContext) (*lsp.CompletionList, error) {
	m := make(map[string]lsp.CompletionItem)
	for _, queriedValuesFiles := range f.Chart.ResolveValueFiles(templateContext.Tail(), f.ChartStore) {
		for _, valuesFile := range queriedValuesFiles.ValuesFiles.AllValuesFiles() {
			values, err := f.DocumentStore.GetValues(valuesFile.URI)
			if err != nil {
				continue
			}
			for _, item := range util.GetValueCompletion(values, queriedValuesFiles.Selector) {
				m[item.InsertText] = item
			}
		}
	}
	completions := []lsp.CompletionItem{}
	for _, item := range m {
		completions = append(completions, item)
	}

	return &lsp.CompletionList{Items: completions, IsIncomplete: false}, nil
}
