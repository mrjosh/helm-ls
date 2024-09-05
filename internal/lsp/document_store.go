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
	templateDocuments sync.Map
	documents         map[string]*sync.Map
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		documents: map[string]*sync.Map{
			("yaml" + ""): new(sync.Map),
			"helm":        new(sync.Map),
		},
		templateDocuments: sync.Map{},
	}
}

func (s *DocumentStore) DidOpen(params *lsp.DidOpenTextDocumentParams, helmlsConfig util.HelmlsConfiguration) (*TemplateDocument, error) {
	logger.Debug(fmt.Sprintf("Opening document %s with langID %s", params.TextDocument.URI, params.TextDocument.LanguageID))

	uri := params.TextDocument.URI
	path := uri.Filename()
	doc := NewTemplateDocument(uri, []byte(params.TextDocument.Text), true, helmlsConfig)
	logger.Debug("Storing doc ", path)
	s.templateDocuments.Store(path, doc)
	return doc, nil
}

func (s *DocumentStore) Store(path string, content []byte, helmlsConfig util.HelmlsConfiguration) {
	_, ok := s.templateDocuments.Load(path)
	if ok {
		return
	}
	fileURI := uri.File(path)
	s.templateDocuments.Store(fileURI.Filename(),
		NewTemplateDocument(fileURI, content, false, helmlsConfig))
}

func (s *DocumentStore) GetTemplateDoc(docuri uri.URI) (*TemplateDocument, bool) {
	path := docuri.Filename()
	d, ok := s.templateDocuments.Load(path)

	if !ok {
		return nil, false
	}
	return d.(*TemplateDocument), ok
}

func (s *DocumentStore) GetAllTemplateDocs() []*TemplateDocument {
	var docs []*TemplateDocument
	s.templateDocuments.Range(func(_, v interface{}) bool {
		docs = append(docs, v.(*TemplateDocument))
		return true
	})
	return docs
}
