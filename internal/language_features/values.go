package languagefeatures

import (
	"fmt"

	lsp "go.lsp.dev/protocol"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
)

type ValuesFeature struct {
	GenericDocumentUseCase
}

func NewValuesFeature(genericDocumentUseCase GenericDocumentUseCase) *ValuesFeature {
	return &ValuesFeature{
		GenericDocumentUseCase: genericDocumentUseCase,
	}
}

func (f *ValuesFeature) References() (result []lsp.Location, err error) {
	includeName, err := f.getTemplateContext()
	if err != nil {
		return []lsp.Location{}, err
	}

	locations := f.getReferenceLocations(includeName)
	return locations, nil
}

func (f *ValuesFeature) getTemplateContext() (lsplocal.TemplateContext, error) {
	templateContext := f.GenericDocumentUseCase.Document.SymbolTable.GetTemplateContext(lsplocal.GetRangeForNode(f.Node))
	if len(templateContext) == 0 || templateContext == nil {
		return lsplocal.TemplateContext{}, fmt.Errorf("no template context found")
	}
	return templateContext, nil
}

func (f *ValuesFeature) getReferenceLocations(templateContext lsplocal.TemplateContext) []lsp.Location {
	locations := []lsp.Location{}
	for _, doc := range f.GenericDocumentUseCase.DocumentStore.GetAllDocs() {
		referenceRanges := doc.SymbolTable.GetValues(templateContext)
		for _, referenceRange := range referenceRanges {
			locations = append(locations, util.RangeToLocation(doc.URI, referenceRange))
		}
	}

	return locations
}
