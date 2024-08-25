package languagefeatures

import (
	lsp "go.lsp.dev/protocol"

	"github.com/mrjosh/helm-ls/internal/charts"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/protocol"
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
)

type IncludesFeature struct {
	*GenericDocumentUseCase
}

type IncludesCallFeature struct {
	*IncludesFeature
}

// should be called on {{ include "name" . }}
func (f *IncludesCallFeature) AppropriateForNode() bool {
	functionCallNode := f.getFunctionCallNode()

	if functionCallNode == nil {
		return false
	}
	_, err := lsplocal.ParseIncludeFunctionCall(functionCallNode, []byte(f.GenericDocumentUseCase.Document.Content))
	return err == nil
}

func (f *IncludesCallFeature) getFunctionCallNode() *sitter.Node {
	var functionCallNode *sitter.Node
	if f.ParentNodeType == gotemplate.NodeTypeArgumentList {
		functionCallNode = f.Node.Parent().Parent()
	}
	if f.ParentNodeType == gotemplate.NodeTypeInterpretedStringLiteral {
		parentParent := f.ParentNode.Parent()
		if parentParent != nil && parentParent.Type() == gotemplate.NodeTypeArgumentList {
			functionCallNode = parentParent.Parent()
		}
	}
	return functionCallNode
}

type IncludesDefinitionFeature struct {
	*IncludesFeature
}

// should be called on {{ define "name" }}
func (f *IncludesDefinitionFeature) AppropriateForNode() bool {
	return f.ParentNodeType == gotemplate.NodeTypeDefineAction && f.NodeType == gotemplate.NodeTypeInterpretedStringLiteral
}

func NewIncludesCallFeature(genericDocumentUseCase *GenericDocumentUseCase) *IncludesCallFeature {
	return &IncludesCallFeature{
		IncludesFeature: &IncludesFeature{genericDocumentUseCase},
	}
}

func NewIncludesDefinitionFeature(genericDocumentUseCase *GenericDocumentUseCase) *IncludesDefinitionFeature {
	return &IncludesDefinitionFeature{
		IncludesFeature: &IncludesFeature{genericDocumentUseCase},
	}
}

func (f *IncludesCallFeature) References() (result []lsp.Location, err error) {
	includeName, err := f.getIncludeName()
	if err != nil {
		return []lsp.Location{}, err
	}

	return f.getReferenceLocations(includeName), nil
}

func (f *IncludesCallFeature) getIncludeName() (string, error) {
	functionCallNode := f.getFunctionCallNode()
	return lsplocal.ParseIncludeFunctionCall(functionCallNode, []byte(f.GenericDocumentUseCase.Document.Content))
}

func (f *IncludesDefinitionFeature) References() (result []lsp.Location, err error) {
	includeName := util.RemoveQuotes(f.GenericDocumentUseCase.NodeContent())
	return f.getReferenceLocations(includeName), nil
}

func (f *IncludesFeature) getReferenceLocations(includeName string) []lsp.Location {
	locations := []lsp.Location{}
	for _, doc := range f.GenericDocumentUseCase.DocumentStore.GetAllDocs() {
		referenceRanges := doc.SymbolTable.GetIncludeReference(includeName)
		for _, referenceRange := range referenceRanges {
			locations = append(locations, util.RangeToLocation(doc.URI, referenceRange))
		}
		if len(locations) > 0 {
			charts.SyncToDisk(doc)
		}
	}

	return locations
}

func (f *IncludesFeature) getDefinitionLocations(includeName string) []lsp.Location {
	locations := []lsp.Location{}
	for _, doc := range f.GenericDocumentUseCase.DocumentStore.GetAllDocs() {
		referenceRanges := doc.SymbolTable.GetIncludeDefinitions(includeName)
		for _, referenceRange := range referenceRanges {
			locations = append(locations, util.RangeToLocation(doc.URI, referenceRange))
		}
		if len(locations) > 0 {
			charts.SyncToDisk(doc)
		}
	}

	return locations
}

func (f *IncludesFeature) getDefinitionsHover(includeName string) protocol.HoverResultsWithFiles {
	result := protocol.HoverResultsWithFiles{}
	for _, doc := range f.GenericDocumentUseCase.DocumentStore.GetAllDocs() {
		referenceRanges := doc.SymbolTable.GetIncludeDefinitions(includeName)
		for _, referenceRange := range referenceRanges {
			node := doc.Ast.RootNode().NamedDescendantForPointRange(referenceRange.StartPoint, referenceRange.EndPoint)
			if node != nil {
				result = append(result, protocol.HoverResultWithFile{
					Value: node.Content([]byte(doc.Content)),
					URI:   doc.URI,
				})
			}
		}
	}

	return result
}

func (f *IncludesCallFeature) Hover() (string, error) {
	includeName, err := f.getIncludeName()
	if err != nil {
		return "", err
	}

	result := f.getDefinitionsHover(includeName)
	return result.FormatHelm(f.GenericDocumentUseCase.Document.URI), nil
}

func (f *IncludesCallFeature) Definition() (result []lsp.Location, err error) {
	includeName, err := f.getIncludeName()
	if err != nil {
		return []lsp.Location{}, err
	}
	return f.getDefinitionLocations(includeName), nil
}

func (f *IncludesCallFeature) Completion() (*lsp.CompletionList, error) {
	items := []lsp.CompletionItem{}
	for _, doc := range f.GenericDocumentUseCase.DocumentStore.GetAllDocs() {
		inlcudeDefinitionNames := doc.SymbolTable.GetAllIncludeDefinitionsNames()
		for _, includeDefinitionName := range inlcudeDefinitionNames {
			items = append(items, lsp.CompletionItem{
				InsertText: includeDefinitionName,
				Kind:       lsp.CompletionItemKindFunction,
				Label:      includeDefinitionName,
			})
		}
	}
	return &lsp.CompletionList{IsIncomplete: false, Items: items}, nil
}
