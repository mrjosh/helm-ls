package languagefeatures

import (
	"fmt"

	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

type GenericTemplateContextFeature struct {
	*GenericDocumentUseCase
}

func (f *GenericTemplateContextFeature) getTemplateContext() (lsplocal.TemplateContext, error) {
	return f.GenericDocumentUseCase.Document.SymbolTable.GetTemplateContext(lsplocal.GetRangeForNode(f.Node))
}

func (f *GenericTemplateContextFeature) getReferencesFromSymbolTable(templateContext lsplocal.TemplateContext) []lsp.Location {
	locations := []lsp.Location{}

	for _, doc := range f.GenericDocumentUseCase.DocumentStore.GetAllDocs() {
		referenceRanges := doc.SymbolTable.GetTemplateContextRanges(templateContext)
		for _, referenceRange := range referenceRanges {
			locations = append(locations, util.RangeToLocation(doc.URI, referenceRange))
		}
	}

	return locations
}

func (f *GenericTemplateContextFeature) builtInOjectDocsLookup(key string, docs []helmdocs.HelmDocumentation) (helmdocs.HelmDocumentation, error) {
	for _, item := range docs {
		if key == item.Name {
			return item, nil
		}
	}

	return helmdocs.HelmDocumentation{}, fmt.Errorf("key <%s> not found on built-in object", key)
}
