package templatehandler

import (
	"context"

	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	templateast "github.com/mrjosh/helm-ls/internal/lsp/template_ast"
	"github.com/mrjosh/helm-ls/internal/protocol"
	lsp "go.lsp.dev/protocol"

	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
)

func (h *TemplateHandler) Completion(ctx context.Context, params *lsp.CompletionParams) (result *lsp.CompletionList, err error) {
	logger.Debug("Running completion with params", params)

	genericDocumentUseCase, err := h.NewGenericDocumentUseCase(params.TextDocumentPositionParams, templateast.NestedNodeAtPositionForCompletion)
	if err != nil {
		return nil, err
	}

	usecases := []languagefeatures.CompletionUseCase{
		languagefeatures.NewTemplateContextFeature(genericDocumentUseCase),
		languagefeatures.NewFunctionCallFeature(genericDocumentUseCase),
		languagefeatures.NewTextFeature(ctx, genericDocumentUseCase, h.yamllsConnector, &params.TextDocumentPositionParams),
		languagefeatures.NewIncludesCallFeature(genericDocumentUseCase),
		languagefeatures.NewVariablesFeature(genericDocumentUseCase),
	}

	for _, usecase := range usecases {
		if usecase.AppropriateForNode() {
			return usecase.Completion()
		}
	}

	// If no usecase matched, we assume we are at {{ }}
	// and provide the basic BuiltInObjects and functions
	items := []helmdocs.HelmDocumentation{}
	for _, v := range helmdocs.BuiltInObjects {
		v.Name = "." + v.Name
		items = append(items, v)
	}

	return protocol.CompletionResults{}.
		WithDocs(items, lsp.CompletionItemKindConstant).
		WithDocs(helmdocs.AllFuncs, lsp.CompletionItemKindFunction).ToList(), nil
}
