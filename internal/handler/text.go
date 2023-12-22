package handler

import (
	"context"
	"encoding/json"
	"errors"

	lspinternal "github.com/mrjosh/helm-ls/internal/lsp"
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

	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}

	var shouldSendFullUpdateToYamlls = false

	// Synchronise changes into the doc's ContentChanges
	doc.ApplyChanges(params.ContentChanges)

	for _, change := range params.ContentChanges {
		var node = lspinternal.NodeAtPosition(doc.Ast, change.Range.Start)
		if node.Type() != "text" {
			shouldSendFullUpdateToYamlls = true
			break
		}
	}

	if shouldSendFullUpdateToYamlls {
		h.yamllsConnector.DocumentDidChangeFullSync(doc, params)
	} else {
		h.yamllsConnector.DocumentDidChange(doc, params)
	}

	return reply(ctx, nil, nil)
}
