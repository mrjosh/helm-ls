package lsp

import (
	"fmt"
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

func (s *DocumentStore) GetAllDocs() []*TemplateDocument {
	var docs []*TemplateDocument
	s.documents.Range(func(_, v interface{}) bool {
		docs = append(docs, v.(*TemplateDocument))
		return true
	})
	return docs
}

func (s *DocumentStore) DidOpen(params *lsp.DidOpenTextDocumentParams, helmlsConfig util.HelmlsConfiguration) (*TemplateDocument, error) {
	logger.Debug(fmt.Sprintf("Opening document %s with langID %s", params.TextDocument.URI, params.TextDocument.LanguageID))

	uri := params.TextDocument.URI
	path := uri.Filename()
	ast := ParseAst(nil, []byte(params.TextDocument.Text))
	doc := &TemplateDocument{
		Document: Document{
			URI:     uri,
			Path:    path,
			Content: []byte(params.TextDocument.Text),
			IsOpen:  true,
		},
		Ast:              ast,
		DiagnosticsCache: NewDiagnosticsCache(helmlsConfig),
		SymbolTable:      NewSymbolTable(ast, []byte(params.TextDocument.Text)),
		IsYaml:           IsYamlDocument(uri, helmlsConfig.YamllsConfiguration),
	}
	logger.Debug("Storing doc ", path)
	s.documents.Store(path, doc)
	return doc, nil
}

func (s *DocumentStore) Store(path string, content []byte, helmlsConfig util.HelmlsConfiguration) {
	_, ok := s.documents.Load(path)
	if ok {
		return
	}
	ast := ParseAst(nil, content)
	fileURI := uri.File(path)
	s.documents.Store(fileURI.Filename(),
		&TemplateDocument{
			Document: Document{
				URI:     fileURI,
				Path:    path,
				Content: content,
				IsOpen:  false,
			},
			Ast:              ast,
			DiagnosticsCache: NewDiagnosticsCache(helmlsConfig),
			SymbolTable:      NewSymbolTable(ast, content),
			IsYaml:           IsYamlDocument(fileURI, helmlsConfig.YamllsConfiguration),
		},
	)
}

func (s *DocumentStore) Get(docuri uri.URI) (*TemplateDocument, bool) {
	path := docuri.Filename()
	d, ok := s.documents.Load(path)

	if !ok {
		return nil, false
	}
	return d.(*TemplateDocument), ok
}

func IsYamlDocument(uri lsp.URI, yamllsConfiguration util.YamllsConfiguration) bool {
	return yamllsConfiguration.EnabledForFilesGlobObject.Match(uri.Filename())
}
