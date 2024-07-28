package languagefeatures

import (
	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	lsp "go.lsp.dev/protocol"
)

type BuiltInObjectsFeature struct {
	*GenericTemplateContextFeature
}

func NewBuiltInObjectsFeature(genericDocumentUseCase *GenericDocumentUseCase) *BuiltInObjectsFeature {
	return &BuiltInObjectsFeature{
		GenericTemplateContextFeature: &GenericTemplateContextFeature{genericDocumentUseCase},
	}
}

func (f *BuiltInObjectsFeature) AppropriateForNode() bool {
	if !(f.ParentNodeType == gotemplate.NodeTypeField && f.NodeType == gotemplate.NodeTypeIdentifier) &&
		f.NodeType != gotemplate.NodeTypeIdentifier &&
		f.NodeType != gotemplate.NodeTypeFieldIdentifier &&
		f.NodeType != gotemplate.NodeTypeDot &&
		f.NodeType != gotemplate.NodeTypeDotSymbol {
		return false
	}

	allowedBuiltIns := []string{}
	for _, allowedBuiltIn := range helmdocs.BuiltInObjects {
		allowedBuiltIns = append(allowedBuiltIns, allowedBuiltIn.Name)
	}

	templateContext, err := f.getTemplateContext()
	if err != nil || len(templateContext) != 1 {
		return false
	}

	for _, allowedBuiltIn := range allowedBuiltIns {
		if templateContext[0] == allowedBuiltIn {
			return true
		}
	}

	return false
}

func (f *BuiltInObjectsFeature) References() (result []lsp.Location, err error) {
	templateContext, err := f.getTemplateContext()
	if err != nil {
		return []lsp.Location{}, err
	}

	locations := f.getReferencesFromSymbolTable(templateContext)

	return append(locations, f.getDefinitionLocations(templateContext)...), err
}

func (f *BuiltInObjectsFeature) getDefinitionLocations(templateContext lsplocal.TemplateContext) []lsp.Location {
	locations := []lsp.Location{}

	switch templateContext[0] {
	case "Values":
		for _, valueFile := range f.Chart.ValuesFiles.AllValuesFiles() {
			locations = append(locations, lsp.Location{URI: valueFile.URI})
		}
		return locations

	case "Chart":
		return []lsp.Location{{URI: f.Chart.ChartMetadata.URI}}
	}

	return []lsp.Location{}
}

func (f *BuiltInObjectsFeature) Hover() (string, error) {
	templateContext, err := f.getTemplateContext()
	if err != nil {
		return "", err
	}

	docs, err := f.builtInOjectDocsLookup(templateContext[0], helmdocs.BuiltInObjects)

	return docs.Doc, err
}

func (f *BuiltInObjectsFeature) Definition() (result []lsp.Location, err error) {
	templateContext, err := f.getTemplateContext()
	if err != nil {
		return []lsp.Location{}, err
	}

	return f.getDefinitionLocations(templateContext), nil
}
