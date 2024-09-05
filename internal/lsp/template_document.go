package lsp

import (
	"strings"

	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
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

// ApplyChanges updates the content of the document from LSP textDocument/didChange events.
func (d *TemplateDocument) ApplyChanges(changes []lsp.TextDocumentContentChangeEvent) {
	d.Document.ApplyChanges(changes)

	d.ApplyChangesToAst(d.Content)
	d.SymbolTable = NewSymbolTable(d.Ast, d.Content)

	d.lines = nil
}

// WordAt returns the word found at the given location.
func (d *TemplateDocument) WordAt(pos lsp.Position) string {
	logger.Debug(pos)

	line, ok := d.getLine(int(pos.Line))
	if !ok {
		return ""
	}
	return util.WordAt(line, int(pos.Character))
}

// getLine returns the line at the given index.
func (d *TemplateDocument) getLine(index int) (string, bool) {
	lines := d.getLines()
	if index < 0 || index > len(lines) {
		return "", false
	}
	return lines[index], true
}

// getLines returns all the lines in the document.
func (d *TemplateDocument) getLines() []string {
	if d.lines == nil {
		// We keep \r on purpose, to avoid messing up position conversions.
		d.lines = strings.Split(string(d.Content), "\n")
	}
	return d.lines
}

// GetContent implements PossibleDependencyFile.
func (d *TemplateDocument) GetContent() []byte {
	return d.Content
}

// GetPath implements PossibleDependencyFile.
func (d *TemplateDocument) GetPath() string {
	return d.Path
}
