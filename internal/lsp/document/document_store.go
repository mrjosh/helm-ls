package document

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/mrjosh/helm-ls/internal/charts"
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

func (s *DocumentStore) DidOpenTemplateDocument(
	params *lsp.DidOpenTextDocumentParams, helmlsConfig util.HelmlsConfiguration,
) (*TemplateDocument, error) {
	uri := params.TextDocument.URI
	path := uri.Filename()
	doc := NewTemplateDocument(uri, []byte(params.TextDocument.Text), true, helmlsConfig)
	logger.Debug("Storing doc ", path)
	s.documents.Store(path, doc)
	return doc, nil
}

func (s *DocumentStore) DidOpenYamlDocument(
	params *lsp.DidOpenTextDocumentParams, helmlsConfig util.HelmlsConfiguration,
) (*YamlDocument, error) {
	uri := params.TextDocument.URI
	path := uri.Filename()
	doc := NewYamlDocument(uri, []byte(params.TextDocument.Text), true, helmlsConfig)
	logger.Debug("Storing doc ", path)
	s.documents.Store(path, doc)
	return doc, nil
}

func (s *DocumentStore) StoreTemplateDocument(path string, content []byte, helmlsConfig util.HelmlsConfiguration) {
	_, ok := s.documents.Load(path)
	if ok {
		return
	}
	fileURI := uri.File(path)
	s.documents.Store(fileURI.Filename(),
		NewTemplateDocument(fileURI, content, false, helmlsConfig))
}

func (s *DocumentStore) GetTemplateDoc(docuri uri.URI) (*TemplateDocument, bool) {
	path := docuri.Filename()
	d, ok := s.documents.Load(path)

	if !ok {
		return nil, false
	}
	doc, ok := d.(*TemplateDocument)
	return doc, ok
}

func (s *DocumentStore) GetYamlDoc(docuri uri.URI) (*YamlDocument, bool) {
	path := docuri.Filename()
	d, ok := s.documents.Load(path)

	if !ok {
		return nil, false
	}

	doc, ok := d.(*YamlDocument)
	return doc, ok
}

func (s *DocumentStore) GetAllTemplateDocs() []*TemplateDocument {
	var docs []*TemplateDocument
	s.documents.Range(func(_, v interface{}) bool {
		doc, ok := v.(*TemplateDocument)
		if !ok {
			return true
		}
		docs = append(docs, doc)
		return true
	})
	return docs
}

func (s *DocumentStore) GetAllYamlDocs() []*YamlDocument {
	var docs []*YamlDocument
	s.documents.Range(func(_, v interface{}) bool {
		doc, ok := v.(*YamlDocument)
		if !ok {
			return true
		}
		docs = append(docs, doc)
		return true
	})
	return docs
}

func (s *DocumentStore) LoadDocsOnNewChart(chart *charts.Chart, helmlsConfig util.HelmlsConfiguration) {
	if chart.HelmChart == nil {
		return
	}

	for _, file := range chart.HelmChart.Templates {
		s.StoreTemplateDocument(filepath.Join(chart.RootURI.Filename(), file.Name), file.Data, helmlsConfig)
	}

	for _, file := range chart.GetDependeciesTemplates() {
		logger.Debug(fmt.Sprintf("Storing dependency %s", file.Path))
		s.StoreTemplateDocument(file.Path, file.Content, helmlsConfig)
	}
}

func (s *DocumentStore) GetDocumentType(uri uri.URI) (DocumentType, bool) {
	path := uri.Filename()
	d, ok := s.documents.Load(path)
	if !ok {
		return DocumentType(""), false
	}
	doc, ok := d.(DocumentInterface)
	return doc.GetDocumentType(), ok
}

func (s *DocumentStore) GetSyncDocument(uri uri.URI) (DocumentInterface, bool) {
	path := uri.Filename()
	d, ok := s.documents.Load(path)
	if !ok {
		return nil, false
	}
	doc, ok := d.(DocumentInterface)
	return doc, ok
}
