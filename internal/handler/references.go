package handler

import (
	"context"

	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) References(ctx context.Context, params *lsp.ReferenceParams) (result []lsp.Location, err error) {
	genericDocumentUseCase, err := h.NewGenericDocumentUseCase(params.TextDocumentPositionParams)
	if err != nil {
		return nil, err
	}

	parentNode := genericDocumentUseCase.Node.Parent()
	pt := parentNode.Type()
	ct := genericDocumentUseCase.Node.Type()

	logger.Println("pt", pt, "ct", ct)
	logger.Println(genericDocumentUseCase.NodeContent())

	if pt == gotemplate.NodeTypeDefineAction && ct == gotemplate.NodeTypeInterpretedStringLiteral {
		includesDefinitionFeature := languagefeatures.NewIncludesDefinitionFeature(genericDocumentUseCase)
		return includesDefinitionFeature.References()
	}

	if pt == gotemplate.NodeTypeArgumentList {
		includesCallFeature := languagefeatures.NewIncludesCallFeature(genericDocumentUseCase)
		return includesCallFeature.References()
	}

	if (pt == gotemplate.NodeTypeField && ct == gotemplate.NodeTypeIdentifier) || ct == gotemplate.NodeTypeFieldIdentifier || ct == gotemplate.NodeTypeField {
		valuesFeature := languagefeatures.NewValuesFeature(genericDocumentUseCase)
		return valuesFeature.References()
	}
	return nil, nil
}
