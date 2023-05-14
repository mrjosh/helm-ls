package yamlls

import (
	"context"

	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector YamllsConnector) DocumentDidOpen(params lsp.DidOpenTextDocumentParams) {
	if yamllsConnector.Conn == nil {
		return
	}

	yamllsConnector.Conn.Notify(context.Background(), lsp.MethodTextDocumentDidOpen, params)
}

func (yamllsConnector YamllsConnector) DocumentDidSave(params lsp.DidSaveTextDocumentParams) {
	if yamllsConnector.Conn == nil {
		return
	}

	yamllsConnector.Conn.Notify(context.Background(), lsp.MethodTextDocumentDidSave, params)
}

func (yamllsConnector YamllsConnector) DocumentDidChange(params lsp.DidChangeTextDocumentParams) {
	if yamllsConnector.Conn == nil {
		return
	}

	yamllsConnector.Conn.Notify(context.Background(), lsp.MethodTextDocumentDidChange, params)
}
