package yamlls

import (
	"context"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector YamllsConnector) DocumentDidOpen(ast *sitter.Tree, params lsp.DidOpenTextDocumentParams) {
	logger.Println("YamllsConnector DocumentDidOpen", params.TextDocument.URI)
	if yamllsConnector.Conn == nil {
		return
	}
	params.TextDocument.Text = trimTemplateForYamllsFromAst(ast, params.TextDocument.Text)

	(*yamllsConnector.Conn).Notify(context.Background(), lsp.MethodTextDocumentDidOpen, params)
}

func (yamllsConnector YamllsConnector) DocumentDidSave(doc *lsplocal.Document, params lsp.DidSaveTextDocumentParams) {
	if yamllsConnector.Conn == nil {
		return
	}
	params.Text = trimTemplateForYamllsFromAst(doc.Ast, doc.Content)

	(*yamllsConnector.Conn).Notify(context.Background(), lsp.MethodTextDocumentDidSave, params)

	yamllsConnector.DocumentDidChangeFullSync(doc, lsp.DidChangeTextDocumentParams{TextDocument: lsp.VersionedTextDocumentIdentifier{
		TextDocumentIdentifier: params.TextDocument,
	},
	})
}

func (yamllsConnector YamllsConnector) DocumentDidChange(doc *lsplocal.Document, params lsp.DidChangeTextDocumentParams) {
	if yamllsConnector.Conn == nil {
		return
	}
	var trimmedText = trimTemplateForYamllsFromAst(doc.Ast, doc.Content)

	logger.Println("Sending DocumentDidChange previous", params)
	for i, change := range params.ContentChanges {
		var (
			start = util.PositionToIndex(change.Range.Start, []byte(doc.Content))
			end   = start + len([]byte(change.Text))
		)

		logger.Println("Start end", start, end)
		logger.Println("Setting change text to ", trimmedText[start:end])
		params.ContentChanges[i].Text = trimmedText[start:end]
	}

	logger.Println("Sending DocumentDidChange", params)
	(*yamllsConnector.Conn).Notify(context.Background(), lsp.MethodTextDocumentDidChange, params)
}

func (yamllsConnector YamllsConnector) DocumentDidChangeFullSync(doc *lsplocal.Document, params lsp.DidChangeTextDocumentParams) {
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
