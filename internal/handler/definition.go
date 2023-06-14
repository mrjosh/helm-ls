package handler

import (
	"context"
	"encoding/json"

	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) handleDefinition(_ context.Context, _ jsonrpc2.Replier, req jsonrpc2.Request) (err error) {

	if req.Params() == nil {
		return &jsonrpc2.Error{Code: jsonrpc2.InvalidParams}
	}

	var params lsp.DefinitionParams
	return json.Unmarshal(req.Params(), &params)
}
