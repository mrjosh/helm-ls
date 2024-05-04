package handler

import (
	"context"

	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/protocol"
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"

	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) Hover(ctx context.Context, params *lsp.HoverParams) (result *lsp.Hover, err error) {
	genericDocumentUseCase, err := h.NewGenericDocumentUseCase(params.TextDocumentPositionParams, lsplocal.NodeAtPosition)
	if err != nil {
		return nil, err
	}

	wordRange := lsplocal.GetLspRangeForNode(genericDocumentUseCase.Node)

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
