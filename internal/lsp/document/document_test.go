package document

import (
	"testing"

	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestDocumentStore(t *testing.T) {
	assert := assert.New(t)

	sut := NewDocumentStore()

	assert.Empty(sut.GetAllTemplateDocs())

	doc, ok := sut.GetTemplateDoc(uri.File("test"))
	assert.Nil(doc)
	assert.False(ok)

	sut.DidOpenTemplateDocument(&protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri.File("test.yaml"),
			LanguageID: "helm",
			Version:    0,
			Text:       "{{ .Values.test }}",
		},
	}, util.DefaultConfig)

	assert.Len(sut.GetAllTemplateDocs(), 1)

	doc, ok = sut.GetTemplateDoc(uri.File("test.yaml"))
	assert.NotNil(doc)
	assert.True(ok)
}

func TestDocumentDocumentTypeForFile(t *testing.T) {
	testCases := []struct {
		fileURI protocol.URI
		langID  protocol.LanguageIdentifier
		want    DocumentType
	}{
		{uri.File("test.yaml"), "helm", TemplateDocumentType},
		{uri.File("somewhere/test.yaml"), "helm", TemplateDocumentType},
		{uri.File("somewhere/test.yaml"), "helm-template", TemplateDocumentType},
		{uri.File("test.yml"), "helm", TemplateDocumentType},
		{uri.File("test.yaml"), "yaml", YamlDocumentType},
		{uri.File("test.yml"), "yaml", YamlDocumentType},
		{uri.File("values.yml"), "helm", YamlDocumentType},
		{uri.File("values.dev.yml"), "helm", YamlDocumentType},
		{uri.File("values.yaml"), "helm", YamlDocumentType},
		{uri.File("values.dev.yaml"), "helm", YamlDocumentType},
		{uri.File("templates/values.dev.yaml"), "helm", TemplateDocumentType},
	}

	for _, testCase := range testCases {
		got := DocumentTypeForFile(testCase.langID, testCase.fileURI)
		assert.Equal(t, testCase.want, got, "DocumentTypeForFile(%s, %s) = %s, want %s", testCase.langID, testCase.fileURI, got, testCase.want)
	}
}

func TestIsValuesYamlFile(t *testing.T) {
	testCases := []struct {
		fileURI protocol.URI
		want    bool
	}{
		{uri.File("values.yaml"), true},
		{uri.File("values.yml"), true},
		{uri.File("values.dev.yaml"), true},
		{uri.File("values.dev.yml"), true},
		{uri.File("values.yaml"), true},
		{uri.File("values.dev.yaml"), true},
		{uri.File("templates/values.dev.yaml"), false},
		{uri.File("deployment.yaml"), false},
		{uri.File("templates/deployment.yaml"), false},
	}

	for _, testCase := range testCases {
		got := IsValuesYamlFile(testCase.fileURI)
		assert.Equal(t, testCase.want, got, "IsValuesYamlFile(%s) = %t, want %t", testCase.fileURI, got, testCase.want)
	}
}
