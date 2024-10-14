package yamlhandler

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/protocol"
)

// DidChange implements handler.LangHandler.
func (h *YamlHandler) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) (err error) {
	panic("unimplemented")
}

// DidOpen implements handler.LangHandler.
func (h *YamlHandler) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams, helmlsConfig util.HelmlsConfiguration) (err error) {
	_, err = h.documents.DidOpenTemplateDocument(params, helmlsConfig)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

// DidSave implements handler.LangHandler.
func (h *YamlHandler) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) (err error) {
	return nil
}
