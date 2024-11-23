package yamlls

import (
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector Connector) InitiallySyncOpenTemplateDocuments(docs []*document.TemplateDocument) {
	if yamllsConnector.server == nil {
		return
	}

	for _, doc := range docs {
		if !doc.IsOpen {
			continue
		}

		doc.IsYaml = yamllsConnector.IsYamllsEnabled(doc.URI)

		if !yamllsConnector.isRelevantFile(doc.URI) {
			continue
		}

		yamllsConnector.DocumentDidOpenTemplate(doc.Ast, lsp.DidOpenTextDocumentParams{
			TextDocument: lsp.TextDocumentItem{
				URI:  doc.URI,
				Text: string(doc.Content),
			},
		})
	}
}

func (yamllsConnector Connector) DocumentDidOpenTemplate(ast *sitter.Tree, params lsp.DidOpenTextDocumentParams) {
	logger.Debug("YamllsConnector DocumentDidOpen", params.TextDocument.URI)

	if !yamllsConnector.shouldRun(params.TextDocument.URI) {
		return
	}
	params.TextDocument.Text = lsplocal.TrimTemplate(ast, []byte(params.TextDocument.Text))

	yamllsConnector.DocumentDidOpen(&params)
}

func (yamllsConnector Connector) DocumentDidSaveTemplate(doc *document.TemplateDocument, params lsp.DidSaveTextDocumentParams) {
	if !yamllsConnector.shouldRun(doc.URI) {
		return
	}

	yamllsConnector.DocumentDidSave(&params)

	// this is required as params.Text has no effect since the default of includeText is false
	yamllsConnector.DocumentDidChangeFullSyncTemplate(doc, lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: params.TextDocument,
		},
	})
}

func (yamllsConnector Connector) DocumentDidChangeTemplate(doc *document.TemplateDocument, params lsp.DidChangeTextDocumentParams) {
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

	yamllsConnector.DocumentDidChange(&params)
}

func (yamllsConnector Connector) DocumentDidChangeFullSyncTemplate(doc *document.TemplateDocument, params lsp.DidChangeTextDocumentParams) {
	if !yamllsConnector.shouldRun(doc.URI) {
		return
	}

	logger.Debug("Sending DocumentDidChange with full sync, current content:", string(doc.Content))
	trimmedText := lsplocal.TrimTemplate(doc.Ast.Copy(), doc.Content)

	params.ContentChanges = []lsp.TextDocumentContentChangeEvent{
		{
			Text: trimmedText,
		},
	}

	logger.Println("Sending DocumentDidChange with full sync", params)
	yamllsConnector.DocumentDidChange(&params)
}

func (yamllsConnector Connector) IsYamllsEnabled(uri lsp.URI) bool {
	return yamllsConnector.EnabledForFilesGlobObject.Match(uri.Filename())
}
