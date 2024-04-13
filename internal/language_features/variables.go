package languagefeatures

import (
	"fmt"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

type VariablesFeature struct {
	*GenericDocumentUseCase
}

func NewVariablesFeature(genericDocumentUseCase *GenericDocumentUseCase) *VariablesFeature {
	return &VariablesFeature{
		GenericDocumentUseCase: genericDocumentUseCase,
	}
}

func (f *VariablesFeature) AppropriateForNode() bool {
	return f.NodeType == gotemplate.NodeTypeIdentifier && f.ParentNodeType == gotemplate.NodeTypeVariable
}

func (f *VariablesFeature) Definition() (result []lsp.Location, err error) {
	variableName := f.GenericDocumentUseCase.NodeContent()
	definitionNode := lsplocal.GetVariableDefinition(variableName, f.GenericDocumentUseCase.ParentNode, f.Document.Content)
	if definitionNode == nil {
		return []lsp.Location{}, fmt.Errorf("Could not find definition for %s. Variable definition not found", variableName)
	}
	return []lsp.Location{{URI: f.Document.URI, Range: lsp.Range{Start: util.PointToPosition(definitionNode.StartPoint())}}}, nil
}
