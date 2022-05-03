package lsp

import (
	"context"
	"encoding/json"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"

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
	case protocol.MethodInitialize:
		return h.handleInitialize(ctx, reply, req)
	case protocol.MethodInitialized:
		return
	case protocol.MethodShutdown:
		return h.handleShutdown(ctx, reply, req)
	case protocol.MethodTextDocumentDidOpen:
		return h.handleTextDocumentDidOpen(ctx, reply, req)
	case protocol.MethodTextDocumentDidClose:
		return h.handleTextDocumentDidClose(ctx, reply, req)
	case protocol.MethodTextDocumentDidChange:
		return h.handleTextDocumentDidChange(ctx, reply, req)
	case protocol.MethodTextDocumentDidSave:
		return h.handleTextDocumentDidSave(ctx, reply, req)
	case protocol.MethodTextDocumentCompletion:
		return h.handleTextDocumentCompletion(ctx, reply, req)
	}

	return jsonrpc2.MethodNotFoundHandler(ctx, reply, req)
}

func (h *langHandler) handleTextDocumentCompletion(_ context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {
	return nil
}

func (h *langHandler) handleInitialize(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {
	var params protocol.InitializeParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	return reply(ctx, protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: protocol.TextDocumentSyncOptions{
				Change:    protocol.TextDocumentSyncKindNone,
				OpenClose: true,
				Save: &protocol.SaveOptions{
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
	var params protocol.DidOpenTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	notification, err := notifcationFromLint(ctx, params.TextDocument.URI)
	return reply(ctx, notification, err)
}

func (h *langHandler) handleTextDocumentDidClose(_ context.Context, _ jsonrpc2.Replier, _ jsonrpc2.Request) (err error) {
	return nil
}

func (h *langHandler) handleTextDocumentDidChange(_ context.Context, _ jsonrpc2.Replier, _ jsonrpc2.Request) (err error) {
	return nil
}

func (h *langHandler) handleTextDocumentDidSave(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {
	var params protocol.DidSaveTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	notification, err := notifcationFromLint(ctx, params.TextDocument.URI)
	return reply(ctx, notification, err)
}
