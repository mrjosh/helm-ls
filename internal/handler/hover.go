package handler

import (
	"context"

	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	lspinternal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"

	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) Hover(ctx context.Context, params *lsp.HoverParams) (result *lsp.Hover, err error) {
	genericDocumentUseCase, err := h.NewGenericDocumentUseCase(params.TextDocumentPositionParams)
	if err != nil {
		return nil, err
	}

	wordRange := lspinternal.GetLspRangeForNode(genericDocumentUseCase.Node)

	usecases := []languagefeatures.HoverUseCase{
		languagefeatures.NewBuiltInObjectsFeature(genericDocumentUseCase), // has to be before template context
		languagefeatures.NewTemplateContextFeature(genericDocumentUseCase),
		languagefeatures.NewIncludesCallFeature(genericDocumentUseCase),
		languagefeatures.NewFunctionCallFeature(genericDocumentUseCase),
	}

	for _, usecase := range usecases {
		if usecase.AppropriateForNode() {
			result, err := usecase.Hover()
			return util.BuildHoverResponse(result, wordRange), err
		}
	}

	if genericDocumentUseCase.NodeType == gotemplate.NodeTypeText {
		word := genericDocumentUseCase.Document.WordAt(params.Position)
		if len(word) > 2 && string(word[len(word)-1]) == ":" {
			word = word[0 : len(word)-1]
		}
		response, err := h.yamllsConnector.CallHover(ctx, *params, word)
		return response, err
	}

	return nil, err
}
