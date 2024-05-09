package lsp

import (
	"fmt"
	"os"
	"sync"

	"github.com/mrjosh/helm-ls/internal/util"
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
		SymbolTable:      NewSymbolTable(ast, []byte(params.TextDocument.Text)),
	}
	logger.Debug("Storing doc ", path)
	s.documents.Store(path, doc)
	return doc, nil
}

func (s *DocumentStore) Store(uri uri.URI, helmlsConfig util.HelmlsConfiguration) error {
	content, err := os.ReadFile(uri.Filename())
	if err != nil {
		logger.Error("Could not open file ", uri.Filename(), " ", err)
		return err
	}

	ast := ParseAst(nil, string(content))
	s.documents.Store(uri.Filename(),
		&Document{
			URI:              uri,
			Path:             uri.Filename(),
			Content:          string(content),
			Ast:              ast,
			DiagnosticsCache: NewDiagnosticsCache(helmlsConfig),
			IsOpen:           false,
			SymbolTable:      NewSymbolTable(ast, content),
		},
	)
	return nil
}

func (s *DocumentStore) Get(docuri uri.URI) (*Document, bool) {
	path := docuri.Filename()
	d, ok := s.documents.Load(path)
	if !ok {
		return nil, false
	}
	return d.(*Document), ok
}
