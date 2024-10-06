package handler

import (
	"context"
	"fmt"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

type LangHandler interface {
	Completion(ctx context.Context, params *lsp.CompletionParams) (result *lsp.CompletionList, err error)
	References(ctx context.Context, params *lsp.ReferenceParams) (result []lsp.Location, err error)
	Hover(ctx context.Context, params *lsp.HoverParams) (result *lsp.Hover, err error)
	Definition(ctx context.Context, params *lsp.DefinitionParams) (result []lsp.Location, err error)
	DocumentSymbol(ctx context.Context, params *lsp.DocumentSymbolParams) (result []interface{}, err error)

	SetChartStore(chartStore *charts.ChartStore)
	SetYamllsConnector(yamllsConnector *yamlls.Connector)
}

func (h *ServerHandler) selectLangHandler(ctx context.Context, uri uri.URI) (LangHandler, error) {
	documentType, ok := h.documents.GetDocumentType(uri)

	if !ok {
		return nil, fmt.Errorf("document %s not found or has invalid language", uri)
	}

	return h.langHandlers[documentType], nil
}
