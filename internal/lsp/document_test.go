package lsp

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
