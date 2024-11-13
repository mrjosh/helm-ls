package yamlls

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/lsp/document"
	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector Connector) InitiallySyncOpenYamlDocuments(docs []*document.YamlDocument) {
	if yamllsConnector.server == nil {
		return
	}

	for _, doc := range docs {
		if !doc.IsOpen {
			continue
		}

		yamllsConnector.DocumentDidOpen(&lsp.DidOpenTextDocumentParams{
			TextDocument: lsp.TextDocumentItem{
				URI:  doc.URI,
				Text: string(doc.Content),
			},
		})
	}
}

func (yamllsConnector Connector) DocumentDidOpen(params *lsp.DidOpenTextDocumentParams) {
	logger.Debug("YamllsConnector DocumentDidOpen", params.TextDocument.URI)
	err := yamllsConnector.server.DidOpen(context.Background(), params)
	if err != nil {
		logger.Error("Error calling yamlls for didOpen", err)
	}
}

func (yamllsConnector Connector) DocumentDidSave(params *lsp.DidSaveTextDocumentParams) {
	err := yamllsConnector.server.DidSave(context.Background(), params)
	if err != nil {
		logger.Error("Error calling yamlls for didSave", err)
	}
}

func (yamllsConnector Connector) DocumentDidChange(params *lsp.DidChangeTextDocumentParams) {
	logger.Debug("Sending DocumentDidChange previous", params)
	err := yamllsConnector.server.DidChange(context.Background(), params)
	if err != nil {
		logger.Println("Error calling yamlls for didChange", err)
	}
}
