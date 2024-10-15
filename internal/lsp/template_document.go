package lsp

import (
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// TemplateDocument represents an helm template file.
type TemplateDocument struct {
	Document
	NeedsRefreshDiagnostics bool
	Ast                     *sitter.Tree
	DiagnosticsCache        DiagnosticsCache
	SymbolTable             *SymbolTable
	IsYaml                  bool
}

func (d *TemplateDocument) GetDocumentType() DocumentType {
	return TemplateDocumentType
}

func NewTemplateDocument(fileURI uri.URI, content []byte, isOpen bool, helmlsConfig util.HelmlsConfiguration) *TemplateDocument {
	ast := ParseAst(nil, content)
	return &TemplateDocument{
		Document:                *NewDocument(fileURI, content, isOpen),
		NeedsRefreshDiagnostics: false,
		Ast:                     ast,
		DiagnosticsCache:        NewDiagnosticsCache(helmlsConfig),
		SymbolTable:             NewSymbolTable(ast, content),
		IsYaml:                  IsYamllsEnabled(fileURI, helmlsConfig.YamllsConfiguration),
	}
}

// ApplyChanges updates the content of the document from LSP textDocument/didChange events.
func (d *TemplateDocument) ApplyChanges(changes []lsp.TextDocumentContentChangeEvent) {
	d.Document.ApplyChanges(changes)

	d.ApplyChangesToAst(d.Content)
	d.SymbolTable = NewSymbolTable(d.Ast, d.Content)
}

func IsYamllsEnabled(uri lsp.URI, yamllsConfiguration util.YamllsConfiguration) bool {
	return yamllsConfiguration.EnabledForFilesGlobObject.Match(uri.Filename())
}

func IsTemplateDocumentLangID(langID lsp.LanguageIdentifier) bool {
	return langID == "helm"
}
