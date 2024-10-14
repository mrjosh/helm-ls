package lsp

import (
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// TemplateDocument represents an helm template file.
type YamlDocument struct {
	Document
}

func (d *YamlDocument) GetDocumentType() DocumentType {
	return YamlDocumentType
}

func NewYamlDocument(fileURI uri.URI, content []byte, isOpen bool, helmlsConfig util.HelmlsConfiguration) *YamlDocument {
	return &YamlDocument{
		Document: *NewDocument(fileURI, content, isOpen),
	}
}

// ApplyChanges updates the content of the document from LSP textDocument/didChange events.
func (d *YamlDocument) ApplyChanges(changes []lsp.TextDocumentContentChangeEvent) {
	d.Document.ApplyChanges(changes)

	d.lines = nil
}
