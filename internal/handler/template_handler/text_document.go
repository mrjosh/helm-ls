package templatehandler

import (
	"context"
	"errors"

	templateast "github.com/mrjosh/helm-ls/internal/lsp/template_ast"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

func (h *TemplateHandler) DidOpen(ctx context.Context, params *lsp.DidOpenTextDocumentParams, helmlsConfig util.HelmlsConfiguration) (err error) {
	doc, err := h.documents.DidOpenTemplateDocument(params, helmlsConfig)
	if err != nil {
		logger.Error(err)
		return err
	}

	h.yamllsConnector.DocumentDidOpenTemplate(doc.Ast, *params)

	return nil
}

func (h *TemplateHandler) DidSave(ctx context.Context, params *lsp.DidSaveTextDocumentParams) (err error) {
	doc, ok := h.documents.GetTemplateDoc(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}

	h.yamllsConnector.DocumentDidSaveTemplate(doc, *params)

	return nil
}

func (h *TemplateHandler) PostDidChange(ctx context.Context, params *lsp.DidChangeTextDocumentParams) (err error) {
	doc, ok := h.documents.GetTemplateDoc(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}

	if doc.Ast == nil {
		return errors.New("Could not get document AST: " + params.TextDocument.URI.Filename())
	}

	shouldSendFullUpdateToYamlls := false
	for _, change := range params.ContentChanges {
		node := templateast.NodeAtPosition(doc.Ast, change.Range.Start)
		if node.Type() != "text" {
			shouldSendFullUpdateToYamlls = true
			break
		}
	}
	if shouldSendFullUpdateToYamlls {
		h.yamllsConnector.DocumentDidChangeFullSyncTemplate(doc, *params)
	} else {
		h.yamllsConnector.DocumentDidChangeTemplate(doc, *params)
	}

	return nil
}
