package languagefeatures

import (
	"fmt"

	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
	symboltable "github.com/mrjosh/helm-ls/internal/lsp/symbol_table"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

type GenericTemplateContextFeature struct {
	*GenericDocumentUseCase
}

func (f *GenericTemplateContextFeature) getTemplateContext() (symboltable.TemplateContext, error) {
	return f.GenericDocumentUseCase.Document.SymbolTable.GetTemplateContext(f.Node.Range())
}

func (f *GenericTemplateContextFeature) getReferencesFromSymbolTable(templateContext symboltable.TemplateContext) []lsp.Location {
	locations := []lsp.Location{}

	for _, doc := range f.GenericDocumentUseCase.DocumentStore.GetAllTemplateDocs() {
		referenceRanges := doc.SymbolTable.GetTemplateContextRanges(templateContext)
		locations = append(locations, util.RangesToLocations(doc.URI, referenceRanges)...)
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
