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
	"github.com/mrjosh/helm-ls/pkg/chart"
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
	nodeContent := f.NodeContent()
	println(nodeContent)

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
		hoverResults = util.HoverResultsWithFiles{}
	)
	for _, valuesFiles := range valuesFiles {
		for _, valuesFile := range valuesFiles.ValuesFiles.AllValuesFiles() {
			result, err := util.GetTableOrValueForSelector(valuesFile.Values, strings.Join(valuesFiles.Selector, "."))
			if err == nil {
				hoverResults = append(hoverResults, util.HoverResultWithFile{URI: valuesFile.URI, Value: result})
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

func (f *TemplateContextFeature) Completion() (result *lsp.CompletionList, err error) {
	templateContext, err := f.getTemplateContext()
	if err != nil {
		return nil, err
	}

	if len(templateContext) == 0 {
		result := helmdocs.BuiltInObjects
		return protocol.NewCompletionResults(result).ToLSP(), nil
	}

	if len(templateContext) == 1 {
		result, ok := helmdocs.BuiltInOjectVals[templateContext[0]]
		if !ok {
			result := helmdocs.BuiltInObjects
			return protocol.NewCompletionResults(result).ToLSP(), nil
		}
		return protocol.NewCompletionResults(result).ToLSP(), nil
	}

	switch templateContext[0] {
	case "Values":
		return f.valuesCompletion(templateContext)
	case "Chart", "Release", "Files", "Capabilities", "Template":
		// TODO: make this more fine, by checking the length
		result, ok := helmdocs.BuiltInOjectVals[templateContext[0]]
		if !ok {
			result := helmdocs.BuiltInObjects
			return protocol.NewCompletionResults(result).ToLSP(), nil
		}
		return protocol.NewCompletionResults(result).ToLSP(), nil

	}

	return nil, nil
}

func (f *TemplateContextFeature) valuesCompletion(templateContext lsplocal.TemplateContext) (*lsp.CompletionList, error) {
	m := make(map[string]lsp.CompletionItem)
	for _, queriedValuesFiles := range f.Chart.ResolveValueFiles(templateContext.Tail(), f.ChartStore) {
		for _, valuesFile := range queriedValuesFiles.ValuesFiles.AllValuesFiles() {
			for _, item := range util.GetValueCompletion(valuesFile.Values, queriedValuesFiles.Selector) {
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
