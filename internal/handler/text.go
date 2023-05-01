package handler

import (
	"context"
	"encoding/json"
	"errors"

	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) handleTextDocumentDidChange(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {

	if req.Params() == nil {
		return &jsonrpc2.Error{Code: jsonrpc2.InvalidParams}
	}

	var params lsp.DidChangeTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	h.yamllsConnector.Conn.Notify(context.Background(), lsp.MethodTextDocumentDidChange, params)

	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}

	// Synchronise changes into the doc's ContentChanges
	doc.ApplyChanges(params.ContentChanges)

	return reply(ctx, nil, nil)
}
