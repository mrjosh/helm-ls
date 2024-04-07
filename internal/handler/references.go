package handler

import (
	"context"

	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) References(ctx context.Context, params *lsp.ReferenceParams) (result []lsp.Location, err error) {
	genericDocumentUseCase, err := h.NewGenericDocumentUseCase(params.TextDocumentPositionParams)
	if err != nil {
		return nil, err
	}

	usecases := []languagefeatures.ReferencesUseCase{
		languagefeatures.NewIncludesDefinitionFeature(genericDocumentUseCase),
		languagefeatures.NewIncludesCallFeature(genericDocumentUseCase),
		languagefeatures.NewValuesFeature(genericDocumentUseCase),
	}

	for _, usecase := range usecases {
		if usecase.AppropriateForNode(genericDocumentUseCase.NodeType, genericDocumentUseCase.ParentNodeType, genericDocumentUseCase.Node) {
			return usecase.References()
		}
	}

	return nil, nil
}
