package handler

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"

	"github.com/mrjosh/helm-ls/internal/log"
)

var logger = log.GetLogger()

type langHandler struct {
	connPool        jsonrpc2.Conn
	linterName      string
	documents       *lsplocal.DocumentStore
	chartStore      *charts.ChartStore
	yamllsConnector *yamlls.Connector
	helmlsConfig    util.HelmlsConfiguration
}

func NewHandler(connPool jsonrpc2.Conn) jsonrpc2.Handler {
	documents := lsplocal.NewDocumentStore()
	handler := &langHandler{
		linterName:      "helm-lint",
		connPool:        connPool,
		documents:       documents,
		helmlsConfig:    util.DefaultConfig,
		yamllsConnector: &yamlls.Connector{},
	}
	logger.Printf("helm-lint-langserver: connections opened")
	return jsonrpc2.ReplyHandler(handler.handle)
}

func (h *langHandler) handle(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	logger.Debug("helm-lint-langserver: request method:", req.Method())

	switch req.Method() {
	case lsp.MethodInitialize:
		return h.handleInitialize(ctx, reply, req)
	case lsp.MethodInitialized:
		go h.retrieveWorkspaceConfiguration(ctx)
		return reply(ctx, nil, nil)
	case lsp.MethodShutdown:
		return h.handleShutdown(ctx, reply, req)
	case lsp.MethodTextDocumentDidOpen:
		return h.handleTextDocumentDidOpen(ctx, reply, req)
	case lsp.MethodTextDocumentDidClose:
		return h.handleTextDocumentDidClose(ctx, reply, req)
	case lsp.MethodTextDocumentDidChange:
		return h.handleTextDocumentDidChange(ctx, reply, req)
	case lsp.MethodTextDocumentDidSave:
		return h.handleTextDocumentDidSave(ctx, reply, req)
	case lsp.MethodTextDocumentCompletion:
		return h.handleTextDocumentCompletion(ctx, reply, req)
	case lsp.MethodTextDocumentDefinition:
		return h.handleDefinition(ctx, reply, req)
	case lsp.MethodTextDocumentHover:
		return h.handleHover(ctx, reply, req)
	case lsp.MethodWorkspaceDidChangeConfiguration:
		return h.handleWorkspaceDidChangeConfiguration(ctx, reply, req)
	case lsp.MethodWorkspaceDidChangeWatchedFiles:
		return h.handleDidChangeWatchedFiles(ctx, reply, req)
	default:
		logger.Debug("Unsupported method", req.Method())
	}

	return jsonrpc2.MethodNotFoundHandler(ctx, reply, req)
}

func (h *langHandler) handleShutdown(_ context.Context, _ jsonrpc2.Replier, _ jsonrpc2.Request) (err error) {
	return h.connPool.Close()
}

func (h *langHandler) handleTextDocumentDidOpen(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {
	var params lsp.DidOpenTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return reply(ctx, nil, err)
	}

	doc, err := h.documents.DidOpen(params, h.helmlsConfig)
	if err != nil {
		logger.Println(err)
		return reply(ctx, nil, err)
	}

	h.yamllsConnector.DocumentDidOpen(doc.Ast, params)

	_, err = h.chartStore.GetChartForDoc(doc.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", doc.URI, err)
	}

	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	chart, err := h.chartStore.GetChartForDoc(doc.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", doc.URI, err)
	}
	notification, err := lsplocal.NotificationFromLint(ctx, h.connPool, chart, doc)
	return reply(ctx, notification, err)
}

func (h *langHandler) handleTextDocumentDidClose(ctx context.Context, reply jsonrpc2.Replier, _ jsonrpc2.Request) (err error) {
	return reply(
		ctx,
		h.connPool.Notify(
			ctx,
			lsp.MethodTextDocumentDidClose,
			nil,
		),
		nil,
	)
}

func (h *langHandler) handleTextDocumentDidSave(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {
	var params lsp.DidSaveTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	chart, err := h.chartStore.GetChartForDoc(doc.URI)
	if err != nil {
		logger.Error("Error getting chart info for file", doc.URI, err)
	}

	h.yamllsConnector.DocumentDidSave(doc, params)
	notification, err := lsplocal.NotificationFromLint(ctx, h.connPool, chart, doc)
	return reply(ctx, notification, err)
}
