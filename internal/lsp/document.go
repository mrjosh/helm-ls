package lsp

import (
	"bytes"
	"strings"

	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
)

// Document represents an opened file.
type Document struct {
	URI                     lsp.DocumentURI
	Path                    string
	NeedsRefreshDiagnostics bool
	Content                 string
	lines                   []string
	Ast                     *sitter.Tree
	DiagnosticsCache        DiagnosticsCache
	IsOpen                  bool
	SymbolTable             *SymbolTable
}

// ApplyChanges updates the content of the document from LSP textDocument/didChange events.
func (d *Document) ApplyChanges(changes []lsp.TextDocumentContentChangeEvent) {
	content := []byte(d.Content)
	for _, change := range changes {
		start, end := util.PositionToIndex(change.Range.Start, content), util.PositionToIndex(change.Range.End, content)

		var buf bytes.Buffer
		buf.Write(content[:start])
		buf.Write([]byte(change.Text))
		buf.Write(content[end:])
		content = buf.Bytes()
	}
	d.Content = string(content)

	d.ApplyChangesToAst(d.Content)
	d.SymbolTable = NewSymbolTable(d.Ast, []byte(d.Content))

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

// getLines returns all the lines in the document.
func (d *Document) getLines() []string {
	if d.lines == nil {
		// We keep \r on purpose, to avoid messing up position conversions.
		d.lines = strings.Split(d.Content, "\n")
	}
	return d.lines
}
