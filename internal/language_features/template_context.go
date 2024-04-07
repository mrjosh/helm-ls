package languagefeatures

import (
	"fmt"

	lsp "go.lsp.dev/protocol"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
)

type TemplateContextFeature struct {
	*GenericDocumentUseCase
}

func NewValuesFeature(genericDocumentUseCase *GenericDocumentUseCase) *TemplateContextFeature {
	return &TemplateContextFeature{
		GenericDocumentUseCase: genericDocumentUseCase,
	}
}

func (f *TemplateContextFeature) AppropriateForNode(currentNodeType string, parentNodeType string, node *sitter.Node) bool {
	return (parentNodeType == gotemplate.NodeTypeField && currentNodeType == gotemplate.NodeTypeIdentifier) || currentNodeType == gotemplate.NodeTypeFieldIdentifier || currentNodeType == gotemplate.NodeTypeField
}

func (f *TemplateContextFeature) References() (result []lsp.Location, err error) {
	includeName, err := f.getTemplateContext()
	if err != nil {
		return []lsp.Location{}, err
	}

	locations := f.getReferenceLocations(includeName)
	return locations, nil
}

func (f *TemplateContextFeature) getTemplateContext() (lsplocal.TemplateContext, error) {
	templateContext := f.GenericDocumentUseCase.Document.SymbolTable.GetTemplateContext(lsplocal.GetRangeForNode(f.Node))
	if len(templateContext) == 0 || templateContext == nil {
		return lsplocal.TemplateContext{}, fmt.Errorf("no template context found")
	}
	return templateContext, nil
}

func (f *TemplateContextFeature) getReferenceLocations(templateContext lsplocal.TemplateContext) []lsp.Location {
	locations := []lsp.Location{}
	for _, doc := range f.GenericDocumentUseCase.DocumentStore.GetAllDocs() {
		referenceRanges := doc.SymbolTable.GetTemplateContextRanges(templateContext)
		for _, referenceRange := range referenceRanges {
			locations = append(locations, util.RangeToLocation(doc.URI, referenceRange))
		}
	}

	return append(locations, f.getDefinitionLocations(templateContext)...)
}

func (f *TemplateContextFeature) getDefinitionLocations(templateContext lsplocal.TemplateContext) []lsp.Location {
	locations := []lsp.Location{}
	if len(templateContext) == 0 || templateContext == nil {
		return []lsp.Location{}
	}

	switch templateContext[0] {
	case "Values":
		for _, value := range f.Chart.ResolveValueFiles(templateContext.Tail(), f.ChartStore) {
			locations = append(locations, value.ValuesFiles.GetPositionsForValue(value.Selector)...)
		}
	}
	return locations
}

func (f *TemplateContextFeature) Hover() (string, error) {
	templateContext, err := f.getTemplateContext()
	return templateContext.Format(), err
}
