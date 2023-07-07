package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"path/filepath"

	"github.com/mrjosh/helm-ls/internal/adapter/fs"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/pkg/chart"
	"github.com/mrjosh/helm-ls/pkg/chartutil"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"

	"github.com/mrjosh/helm-ls/internal/log"
)

var logger = log.GetLogger()

type langHandler struct {
	connPool      jsonrpc2.Conn
	linterName    string
	documents     *lsplocal.DocumentStore
	values        chartutil.Values
	chartMetadata chart.Metadata
}

func NewHandler(connPool jsonrpc2.Conn) jsonrpc2.Handler {
	fileStorage, _ := fs.NewFileStorage("")
	handler := &langHandler{
		linterName: "helm-lint",
		connPool:   connPool,
		documents:  lsplocal.NewDocumentStore(fileStorage),
		values:     make(map[string]interface{}),
	}
	logger.Printf("helm-lint-langserver: connections opened")
	return jsonrpc2.ReplyHandler(handler.handle)
}

func (h *langHandler) handle(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	logger.Debug("helm-lint-langserver: request:", req)

	switch req.Method() {
	case lsp.MethodInitialize:
		return h.handleInitialize(ctx, reply, req)
	case lsp.MethodInitialized:
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
	}

	return jsonrpc2.MethodNotFoundHandler(ctx, reply, req)
}

func (h *langHandler) handleInitialize(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {

	var params lsp.InitializeParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	if len(params.WorkspaceFolders) == 0 {
		return errors.New("length WorkspaceFolders is 0")
	}

	workspace_uri, err := url.Parse(params.WorkspaceFolders[0].URI)
	if err != nil {
		return err
	}

	vf := filepath.Join(workspace_uri.Path, "values.yaml")
	vals, err := chartutil.ReadValuesFile(vf)
	if err != nil {
		logger.Println("Error loading values.yaml file", err)
		return err
	}
	h.values = vals

	chartFile := filepath.Join(workspace_uri.Path, "Chart.yaml")
	chartMetadata, err := chartutil.LoadChartfile(chartFile)
	if err != nil {
		logger.Println("Error loading Chart.yaml file", err)
		return err
	}
	h.chartMetadata = *chartMetadata

	return reply(ctx, lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: lsp.TextDocumentSyncOptions{
				Change:    lsp.TextDocumentSyncKindFull,
				OpenClose: true,
				Save: &lsp.SaveOptions{
					IncludeText: true,
				},
			},
			CompletionProvider: &lsp.CompletionOptions{
				TriggerCharacters: []string{".", "$."},
				ResolveProvider:   false,
			},
			HoverProvider: true,
		},
	}, nil)
}

func (h *langHandler) handleShutdown(_ context.Context, _ jsonrpc2.Replier, _ jsonrpc2.Request) (err error) {
	return h.connPool.Close()
}

func (h *langHandler) handleTextDocumentDidOpen(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {

	var params lsp.DidOpenTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return reply(ctx, nil, err)
	}

	if _, err = h.documents.DidOpen(params); err != nil {
		logger.Println(err)
		return reply(ctx, nil, err)
	}

	notification, err := lsplocal.NotifcationFromLint(ctx, h.connPool, params.TextDocument.URI)
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

	notification, err := lsplocal.NotifcationFromLint(ctx, h.connPool, params.TextDocument.URI)
	return reply(ctx, notification, err)
}
