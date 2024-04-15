package languagefeatures

import (
	"fmt"

	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
	"github.com/mrjosh/helm-ls/internal/protocol"
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	lsp "go.lsp.dev/protocol"
)

type FunctionCallFeature struct {
	*GenericDocumentUseCase
}

func NewFunctionCallFeature(genericDocumentUseCase *GenericDocumentUseCase) *FunctionCallFeature {
	return &FunctionCallFeature{
		GenericDocumentUseCase: genericDocumentUseCase,
	}
}

func (f *FunctionCallFeature) AppropriateForNode() bool {
	return f.NodeType == gotemplate.NodeTypeIdentifier && f.ParentNodeType == gotemplate.NodeTypeFunctionCall
}

func (f *FunctionCallFeature) Hover() (string, error) {
	documentation, ok := helmdocs.GetFunctionByName(f.NodeContent())
	if !ok {
		return "", fmt.Errorf("could not find documentation for function %s", f.NodeContent())
	}
	return fmt.Sprintf("%s\n\n%s", documentation.Detail, documentation.Doc), nil
}

func (f *FunctionCallFeature) Completion() (result *lsp.CompletionList, err error) {
	return protocol.NewCompletionResults(helmdocs.AllFuncs).ToLSP(), nil
}
