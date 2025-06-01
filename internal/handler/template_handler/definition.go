package templatehandler

import (
	"context"

	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	templateast "github.com/mrjosh/helm-ls/internal/lsp/template_ast"
	lsp "go.lsp.dev/protocol"
)

func (h *TemplateHandler) Definition(_ context.Context, params *lsp.DefinitionParams) (result []lsp.Location, err error) {
	genericDocumentUseCase, err := h.NewGenericDocumentUseCase(params.TextDocumentPositionParams, templateast.NodeAtPosition)
	if err != nil {
		return nil, err
	}

	usecases := []languagefeatures.DefinitionUseCase{
		languagefeatures.NewBuiltInObjectsFeature(genericDocumentUseCase), // has to be before template context
		languagefeatures.NewVariablesFeature(genericDocumentUseCase),
		languagefeatures.NewTemplateContextFeature(genericDocumentUseCase),
		languagefeatures.NewIncludesCallFeature(genericDocumentUseCase),
	}

	for _, usecase := range usecases {
		if usecase.AppropriateForNode() {
			result, err := usecase.Definition()
			return result, err
		}
	}

	return nil, nil
}
