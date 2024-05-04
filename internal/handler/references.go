package handler

import (
	"context"

	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) References(_ context.Context, params *lsp.ReferenceParams) (result []lsp.Location, err error) {
	genericDocumentUseCase, err := h.NewGenericDocumentUseCase(params.TextDocumentPositionParams, lsplocal.NodeAtPosition)
	if err != nil {
		return nil, err
	}

	usecases := []languagefeatures.ReferencesUseCase{
		languagefeatures.NewIncludesDefinitionFeature(genericDocumentUseCase),
		languagefeatures.NewIncludesCallFeature(genericDocumentUseCase),
		languagefeatures.NewTemplateContextFeature(genericDocumentUseCase),
	}

	for _, usecase := range usecases {
		if usecase.AppropriateForNode() {
			return usecase.References()
		}
	}

	return nil, nil
}
