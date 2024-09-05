package yamlls

import (
	"context"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector Connector) InitiallySyncOpenDocuments(docs []*lsplocal.TemplateDocument) {
	if yamllsConnector.server == nil {
		return
	}

	for _, doc := range docs {
		if !doc.IsOpen {
			continue
		}

		doc.IsYaml = lsplocal.IsYamlDocument(doc.URI, yamllsConnector.config)
		if !yamllsConnector.isRelevantFile(doc.URI) {
			continue
		}

		yamllsConnector.DocumentDidOpen(doc.Ast, lsp.DidOpenTextDocumentParams{
			TextDocument: lsp.TextDocumentItem{
				URI:  doc.URI,
				Text: string(doc.Content),
			},
		})
	}
}

func (yamllsConnector Connector) DocumentDidOpen(ast *sitter.Tree, params lsp.DidOpenTextDocumentParams) {
	logger.Debug("YamllsConnector DocumentDidOpen", params.TextDocument.URI)

	if !yamllsConnector.shouldRun(params.TextDocument.URI) {
		return
	}
	params.TextDocument.Text = lsplocal.TrimTemplate(ast, []byte(params.TextDocument.Text))

	err := yamllsConnector.server.DidOpen(context.Background(), &params)
	if err != nil {
		logger.Error("Error calling yamlls for didOpen", err)
	}
}

func (yamllsConnector Connector) DocumentDidSave(doc *lsplocal.TemplateDocument, params lsp.DidSaveTextDocumentParams) {
	if !yamllsConnector.shouldRun(doc.URI) {
		return
	}

	params.Text = lsplocal.TrimTemplate(doc.Ast, doc.Content)

	err := yamllsConnector.server.DidSave(context.Background(), &params)
	if err != nil {
		logger.Error("Error calling yamlls for didSave", err)
	}

	yamllsConnector.DocumentDidChangeFullSync(doc, lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: params.TextDocument,
		},
	})
}

func (yamllsConnector Connector) DocumentDidChange(doc *lsplocal.TemplateDocument, params lsp.DidChangeTextDocumentParams) {
	if !yamllsConnector.shouldRun(doc.URI) {
		return
	}
	trimmedText := lsplocal.TrimTemplate(doc.Ast, doc.Content)

	logger.Debug("Sending DocumentDidChange previous", params)
	for i, change := range params.ContentChanges {
		var (
			start = util.PositionToIndex(change.Range.Start, []byte(doc.Content))
			end   = start + len([]byte(change.Text))
		)

		if end >= len(trimmedText) {
			end = len(trimmedText) - 1
		}

		logger.Debug("Start end", start, end)
		logger.Debug("Setting change text to ", trimmedText[start:end])
		params.ContentChanges[i].Text = trimmedText[start:end]
	}

	logger.Debug("Sending DocumentDidChange", params)
	err := yamllsConnector.server.DidChange(context.Background(), &params)
	if err != nil {
		logger.Println("Error calling yamlls for didChange", err)
	}
}

func (yamllsConnector Connector) DocumentDidChangeFullSync(doc *lsplocal.TemplateDocument, params lsp.DidChangeTextDocumentParams) {
	if !yamllsConnector.shouldRun(doc.URI) {
		return
	}

	logger.Debug("Sending DocumentDidChange with full sync, current content:", doc.Content)
	trimmedText := lsplocal.TrimTemplate(doc.Ast.Copy(), doc.Content)

	params.ContentChanges = []lsp.TextDocumentContentChangeEvent{
		{
			Text: trimmedText,
		},
	}

	logger.Println("Sending DocumentDidChange with full sync", params)
	err := yamllsConnector.server.DidChange(context.Background(), &params)
	if err != nil {
		logger.Println("Error calling yamlls for didChange", err)
	}
}
