package lsp

import (
	"context"
	"encoding/json"

	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"

	"github.com/mrjosh/helm-lint-ls/internal/log"
)

var logger = log.GetLogger()

func NewHandler(connPool jsonrpc2.Conn) jsonrpc2.Handler {
	handler := &langHandler{
		linterName: "helm-lint",
		connPool:   connPool,
	}
	logger.Printf("helm-lint-langserver: connections opened")
	return jsonrpc2.ReplyHandler(handler.handle)
}

type langHandler struct {
	connPool   jsonrpc2.Conn
	linterName string
}

func (h *langHandler) handle(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {
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
	}

	return jsonrpc2.MethodNotFoundHandler(ctx, reply, req)
}

func (h *langHandler) handleTextDocumentCompletion(_ context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {
	return nil
}

func (h *langHandler) handleInitialize(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {
	var params lsp.InitializeParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	return reply(ctx, lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: lsp.TextDocumentSyncOptions{
				Change:    lsp.TextDocumentSyncKindNone,
				OpenClose: true,
				Save: &lsp.SaveOptions{
					IncludeText: true,
				},
			},
		},
	}, nil)
}

func (h *langHandler) handleShutdown(_ context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {
	return h.connPool.Close()
}

func (h *langHandler) handleTextDocumentDidOpen(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {

	var params lsp.DidOpenTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	notification, err := notifcationFromLint(ctx, h.connPool, params.TextDocument.URI)
	return reply(ctx, notification, err)
}

func (h *langHandler) handleTextDocumentDidClose(_ context.Context, _ jsonrpc2.Replier, _ jsonrpc2.Request) (err error) {
	return nil
}

func (h *langHandler) handleTextDocumentDidChange(_ context.Context, _ jsonrpc2.Replier, _ jsonrpc2.Request) (err error) {
	return nil
}

func (h *langHandler) handleTextDocumentDidSave(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {
	var params lsp.DidSaveTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	notification, err := notifcationFromLint(ctx, h.connPool, params.TextDocument.URI)
	return reply(ctx, notification, err)
}
