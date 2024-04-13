package handler

import (
	"context"

	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) Definition(_ context.Context, params *lsp.DefinitionParams) (result []lsp.Location, err error) {
	genericDocumentUseCase, err := h.NewGenericDocumentUseCase(params.TextDocumentPositionParams)
	if err != nil {
		return nil, err
	}
	usecases := []languagefeatures.DefinitionUseCase{
		languagefeatures.NewBuiltInObjectsFeature(genericDocumentUseCase), // has to be before template context
		languagefeatures.NewTemplateContextFeature(genericDocumentUseCase),
		languagefeatures.NewIncludesCallFeature(genericDocumentUseCase),
		languagefeatures.NewVariablesFeature(genericDocumentUseCase),
	}

	for _, usecase := range usecases {
		if usecase.AppropriateForNode() {
			result, err := usecase.Definition()
			return result, err
		}
	}

	return nil, nil
}
