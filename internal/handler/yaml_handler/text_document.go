package yamlhandler

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/protocol"
)

// DidOpen implements handler.LangHandler.
func (h *YamlHandler) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams, helmlsConfig util.HelmlsConfiguration) (err error) {
	h.yamllsConnector.DocumentDidOpen(params)
	_, err = h.documents.DidOpenYamlDocument(params, helmlsConfig)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

// DidSave implements handler.LangHandler.
func (h *YamlHandler) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) (err error) {
	h.yamllsConnector.DocumentDidSave(params)
	return nil
}

// DidChange implements handler.LangHandler.
func (h *YamlHandler) PostDidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) (err error) {
	h.yamllsConnector.DocumentDidChange(params)
	return nil
}
