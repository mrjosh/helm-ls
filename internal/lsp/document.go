package lsp

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// documentStore holds opened documents.
type DocumentStore struct {
	documents sync.Map
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		documents: sync.Map{},
	}
}

func (s *DocumentStore) GetAllDocs() []*Document {
	var docs []*Document
	s.documents.Range(func(_, v interface{}) bool {
		docs = append(docs, v.(*Document))
		return true
	})
	return docs
}

func (s *DocumentStore) DidOpen(params *lsp.DidOpenTextDocumentParams, helmlsConfig util.HelmlsConfiguration) (*Document, error) {
	logger.Debug(fmt.Sprintf("Opening document %s with langID %s", params.TextDocument.URI, params.TextDocument.LanguageID))

	uri := params.TextDocument.URI
	path := uri.Filename()
	ast := ParseAst(nil, params.TextDocument.Text)
	doc := &Document{
		URI:              uri,
		Path:             path,
		Content:          params.TextDocument.Text,
		Ast:              ast,
		DiagnosticsCache: NewDiagnosticsCache(helmlsConfig),
		IsOpen:           true,
		SymbolTable:      NewSymbolTable(ast),
	}
	logger.Debug("Storing doc ", path)
	s.documents.Store(path, doc)
	return doc, nil
}

func (s *DocumentStore) Get(docuri uri.URI) (*Document, bool) {
	path := docuri.Filename()
	d, ok := s.documents.Load(path)
	if !ok {
		return nil, false
	}
	return d.(*Document), ok
}

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

	d.lines = nil
}

// WordAt returns the word found at the given location.
func (d *Document) WordAt(pos lsp.Position) string {
	logger.Debug(pos)

	line, ok := d.GetLine(int(pos.Line))
	if !ok {
		return ""
	}
	return util.WordAt(line, int(pos.Character))
}

// GetLine returns the line at the given index.
func (d *Document) GetLine(index int) (string, bool) {
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
