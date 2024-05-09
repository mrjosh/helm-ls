package languagefeatures

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/documentation/godocs"
	"github.com/mrjosh/helm-ls/internal/protocol"
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	lsp "go.lsp.dev/protocol"
)

type TextFeature struct {
	*GenericDocumentUseCase
	textDocumentPosition *lsp.TextDocumentPositionParams
	ctx                  context.Context
	yamllsConnector      *yamlls.Connector
}

func NewTextFeature(
	ctx context.Context,
	genericDocumentUseCase *GenericDocumentUseCase,
	yamllsConnector *yamlls.Connector,
	textDocumentPosition *lsp.TextDocumentPositionParams,
) *TextFeature {
	return &TextFeature{
		GenericDocumentUseCase: genericDocumentUseCase,
		textDocumentPosition:   textDocumentPosition,
		ctx:                    ctx,
		yamllsConnector:        yamllsConnector,
	}
}

func (f *TextFeature) AppropriateForNode() bool {
	return f.NodeType == gotemplate.NodeTypeText || f.NodeType == gotemplate.NodeTypeTemplate
}

func (f *TextFeature) Completion() (result *lsp.CompletionList, err error) {
	comletions := f.yamllsCompletions(&lsp.CompletionParams{
		TextDocumentPositionParams: *f.textDocumentPosition,
	})

	return protocol.CompletionResults{Items: comletions}.WithSnippets(godocs.TextSnippets).ToList(), nil
}

func (f *TextFeature) yamllsCompletions(params *lsp.CompletionParams) []lsp.CompletionItem {
	response, err := f.yamllsConnector.CallCompletion(f.ctx, params)
	if err != nil {
		logger.Error("Error getting yamlls completions", err)
		return []lsp.CompletionItem{}
	}
	logger.Debug("Got completions from yamlls", response)
	return response.Items
}
