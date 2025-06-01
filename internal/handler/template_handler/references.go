package templatehandler

import (
	"context"

	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	templateast "github.com/mrjosh/helm-ls/internal/lsp/template_ast"
	lsp "go.lsp.dev/protocol"
)

func (h *TemplateHandler) References(_ context.Context, params *lsp.ReferenceParams) (result []lsp.Location, err error) {
	genericDocumentUseCase, err := h.NewGenericDocumentUseCase(params.TextDocumentPositionParams, templateast.NodeAtPosition)
	if err != nil {
		return nil, err
	}

	usecases := []languagefeatures.ReferencesUseCase{
		languagefeatures.NewIncludesDefinitionFeature(genericDocumentUseCase),
		languagefeatures.NewIncludesCallFeature(genericDocumentUseCase),
		languagefeatures.NewTemplateContextFeature(genericDocumentUseCase),
		languagefeatures.NewVariablesFeature(genericDocumentUseCase),
	}

	for _, usecase := range usecases {
		if usecase.AppropriateForNode() {
			return usecase.References()
		}
	}

	return nil, nil
}
