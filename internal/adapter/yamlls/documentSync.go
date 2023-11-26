package yamlls

import (
	"context"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector Connector) InitiallySyncOpenDocuments(docs []*lsplocal.Document) {
	if yamllsConnector.Conn == nil {
		return
	}
	for _, doc := range docs {
		yamllsConnector.DocumentDidOpen(doc.Ast, lsp.DidOpenTextDocumentParams{
			TextDocument: lsp.TextDocumentItem{
				URI:  doc.URI,
				Text: doc.Content,
			},
		})
	}
}

func (yamllsConnector Connector) DocumentDidOpen(ast *sitter.Tree, params lsp.DidOpenTextDocumentParams) {
	logger.Println("YamllsConnector DocumentDidOpen", params.TextDocument.URI)
	if yamllsConnector.Conn == nil {
		return
	}
	params.TextDocument.Text = trimTemplateForYamllsFromAst(ast, params.TextDocument.Text)

	err := (*yamllsConnector.Conn).Notify(context.Background(), lsp.MethodTextDocumentDidOpen, params)
	if err != nil {
		logger.Println("Error calling yamlls for didOpen", err)
	}
}

func (yamllsConnector Connector) DocumentDidSave(doc *lsplocal.Document, params lsp.DidSaveTextDocumentParams) {
	if yamllsConnector.Conn == nil {
		return
	}
	params.Text = trimTemplateForYamllsFromAst(doc.Ast, doc.Content)

	err := (*yamllsConnector.Conn).Notify(context.Background(), lsp.MethodTextDocumentDidSave, params)
	if err != nil {
		logger.Println("Error calling yamlls for didSave", err)
	}

	yamllsConnector.DocumentDidChangeFullSync(doc, lsp.DidChangeTextDocumentParams{TextDocument: lsp.VersionedTextDocumentIdentifier{
		TextDocumentIdentifier: params.TextDocument,
	},
	})
}

func (yamllsConnector Connector) DocumentDidChange(doc *lsplocal.Document, params lsp.DidChangeTextDocumentParams) {
	if yamllsConnector.Conn == nil {
		return
	}
	var trimmedText = trimTemplateForYamllsFromAst(doc.Ast, doc.Content)

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
	err := (*yamllsConnector.Conn).Notify(context.Background(), lsp.MethodTextDocumentDidChange, params)
	if err != nil {
		logger.Println("Error calling yamlls for didChange", err)
	}
}

func (yamllsConnector Connector) DocumentDidChangeFullSync(doc *lsplocal.Document, params lsp.DidChangeTextDocumentParams) {
	if yamllsConnector.Conn == nil {
		return
	}

	logger.Println("Sending DocumentDidChange with full sync, current content:", doc.Content)
	var trimmedText = trimTemplateForYamllsFromAst(doc.Ast.Copy(), doc.Content)

	params.ContentChanges = []lsp.TextDocumentContentChangeEvent{
		{
			Text: trimmedText,
		},
	}

	logger.Println("Sending DocumentDidChange with full sync", params)
	(*yamllsConnector.Conn).Notify(context.Background(), lsp.MethodTextDocumentDidChange, params)
}
