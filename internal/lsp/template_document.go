package lsp

import (
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// TemplateDocument represents an opened file.
type TemplateDocument struct {
	Document
	NeedsRefreshDiagnostics bool
	Ast                     *sitter.Tree
	DiagnosticsCache        DiagnosticsCache
	SymbolTable             *SymbolTable
	IsYaml                  bool
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

	d.lines = nil
}

// WordAt returns the word found at the given location.
func (d *Document) WordAt(pos lsp.Position) string {
	logger.Debug(pos)

	line, ok := d.getLine(int(pos.Line))
	if !ok {
		return ""
	}
	return util.WordAt(line, int(pos.Character))
}

// getLine returns the line at the given index.
func (d *Document) getLine(index int) (string, bool) {
	lines := d.getLines()
	if index < 0 || index > len(lines) {
		return "", false
	}
	return lines[index], true
}

func IsYamllsEnabled(uri lsp.URI, yamllsConfiguration util.YamllsConfiguration) bool {
	return yamllsConfiguration.EnabledForFilesGlobObject.Match(uri.Filename())
}

func IsTemplateDocumentLangID(langID lsp.LanguageIdentifier) bool {
	return langID == "helm"
}
