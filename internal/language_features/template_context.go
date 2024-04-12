package languagefeatures

import (
	"fmt"
	"reflect"
	"strings"

	lsp "go.lsp.dev/protocol"

	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/protocol"
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	"helm.sh/helm/v3/pkg/chart"
)

type TemplateContextFeature struct {
	*GenericTemplateContextFeature
}

func NewTemplateContextFeature(genericDocumentUseCase *GenericDocumentUseCase) *TemplateContextFeature {
	return &TemplateContextFeature{
		GenericTemplateContextFeature: &GenericTemplateContextFeature{genericDocumentUseCase},
	}
}

func (f *TemplateContextFeature) AppropriateForNode() bool {
	if f.NodeType == gotemplate.NodeTypeDot || f.NodeType == gotemplate.NodeTypeDotSymbol {
		return true
	}
	return (f.ParentNodeType == gotemplate.NodeTypeField && f.NodeType == gotemplate.NodeTypeIdentifier) ||
		f.NodeType == gotemplate.NodeTypeFieldIdentifier ||
		f.NodeType == gotemplate.NodeTypeField
}

func (f *TemplateContextFeature) References() (result []lsp.Location, err error) {
	templateContext, err := f.getTemplateContext()
	if err != nil {
		return []lsp.Location{}, err
	}

	locations := f.getReferenceLocations(templateContext)
	return append(locations, f.getDefinitionLocations(templateContext)...), nil
}

func (f *TemplateContextFeature) Definition() (result []lsp.Location, err error) {
	templateContext, err := f.getTemplateContext()
	if err != nil {
		return []lsp.Location{}, err
	}
	return f.getDefinitionLocations(templateContext), nil
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

	switch templateContext[0] {
	case "Values":
		for _, value := range f.Chart.ResolveValueFiles(templateContext.Tail(), f.ChartStore) {
			locations = append(locations, value.ValuesFiles.GetPositionsForValue(value.Selector)...)
		}
		return locations
	case "Chart":
		location, _ := f.Chart.GetValueLocation(templateContext.Tail())
		return []lsp.Location{location}
	}
	return locations
}

func (f *TemplateContextFeature) Hover() (string, error) {
	templateContext, err := f.getTemplateContext()

	switch templateContext[0] {
	case "Values":
		return f.valuesHover(templateContext.Tail())
	case "Chart":
		docs, err := f.builtInOjectDocsLookup(templateContext.Tail().Format(), helmdocs.BuiltInOjectVals[templateContext[0]])
		value := f.getMetadataField(&f.Chart.ChartMetadata.Metadata, docs.Name)
		return fmt.Sprintf("%s\n\n%s\n", docs.Doc, value), err
	case "Release", "Files", "Capabilities", "Template":
		docs, err := f.builtInOjectDocsLookup(templateContext.Tail().Format(), helmdocs.BuiltInOjectVals[templateContext[0]])
		return docs.Doc, err
	}

	return templateContext.Format(), err
}

func (f *TemplateContextFeature) valuesHover(templateContext lsplocal.TemplateContext) (string, error) {
	var (
		valuesFiles  = f.Chart.ResolveValueFiles(templateContext, f.ChartStore)
		hoverResults = protocol.HoverResultsWithFiles{}
	)
	for _, valuesFiles := range valuesFiles {
		for _, valuesFile := range valuesFiles.ValuesFiles.AllValuesFiles() {
			result, err := util.GetTableOrValueForSelector(valuesFile.Values, strings.Join(valuesFiles.Selector, "."))
			if err == nil {
				hoverResults = append(hoverResults, protocol.HoverResultWithFile{URI: valuesFile.URI, Value: result})
			}
		}
	}
	return hoverResults.Format(f.ChartStore.RootURI), nil
}

func (f *TemplateContextFeature) getMetadataField(v *chart.Metadata, fieldName string) string {
	r := reflect.ValueOf(v)
	field := reflect.Indirect(r).FieldByName(fieldName)
	return util.FormatToYAML(field, fieldName)
}
