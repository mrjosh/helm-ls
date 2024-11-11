package templatehandler

import (
	"context"

	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	templateast "github.com/mrjosh/helm-ls/internal/lsp/template_ast"
	"github.com/mrjosh/helm-ls/internal/protocol"
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"

	lsp "go.lsp.dev/protocol"
)

func (h *TemplateHandler) Hover(ctx context.Context, params *lsp.HoverParams) (result *lsp.Hover, err error) {
	genericDocumentUseCase, err := h.NewGenericDocumentUseCase(params.TextDocumentPositionParams, templateast.NodeAtPosition)
	if err != nil {
		return nil, err
	}

	wordRange := templateast.GetLspRangeForNode(genericDocumentUseCase.Node)

	usecases := []languagefeatures.HoverUseCase{
		languagefeatures.NewBuiltInObjectsFeature(genericDocumentUseCase), // has to be before template context
		languagefeatures.NewTemplateContextFeature(genericDocumentUseCase),
		languagefeatures.NewIncludesCallFeature(genericDocumentUseCase),
		languagefeatures.NewFunctionCallFeature(genericDocumentUseCase),
	}

	for _, usecase := range usecases {
		if usecase.AppropriateForNode() {
			result, err := usecase.Hover()
			return protocol.BuildHoverResponse(result, wordRange), err
		}
	}

	if genericDocumentUseCase.NodeType == gotemplate.NodeTypeText {
		word := genericDocumentUseCase.Document.WordAt(params.Position)
		response, err := h.yamllsConnector.CallHover(ctx, *params, word)
		return response, err
	}

	return nil, err
}
