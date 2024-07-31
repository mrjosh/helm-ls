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

	assert.Empty(sut.GetAllDocs())

	doc, ok := sut.Get(uri.File("test"))
	assert.Nil(doc)
	assert.False(ok)

	sut.DidOpen(&protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri.File("test.yaml"),
			LanguageID: "helm",
			Version:    0,
			Text:       "{{ .Values.test }}",
		},
	}, util.DefaultConfig)

	assert.Len(sut.GetAllDocs(), 1)

	doc, ok = sut.Get(uri.File("test.yaml"))
	assert.NotNil(doc)
	assert.True(ok)
}

func TestApplyChanges(t *testing.T) {
	assert := assert.New(t)

	documentStore := NewDocumentStore()
	documentStore.DidOpen(&protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri.File("test.yaml"),
			LanguageID: "helm",
			Text:       `{{ .Values.test }}`,
		},
	}, util.DefaultConfig)

	doc, ok := documentStore.Get(uri.File("test.yaml"))
	assert.True(ok)
	assert.Equal("{{ .Values.test }}", doc.Content)

	doc.ApplyChanges([]protocol.TextDocumentContentChangeEvent{
		{Range: protocol.Range{Start: protocol.Position{Line: 0, Character: 18}, End: protocol.Position{Line: 0, Character: 18}}, Text: "\n"},
		{Range: protocol.Range{Start: protocol.Position{Line: 1, Character: 0}, End: protocol.Position{Line: 1, Character: 0}}, Text: "\n"},
		{Range: protocol.Range{Start: protocol.Position{Line: 1, Character: 0}, End: protocol.Position{Line: 1, Character: 0}}, Text: "spec:\n  replicas: {{ .Values.replicaCount }}\n  selector:\n    matchLabels:\n      {{- include \"hello-world.selectorLabels\" . | nindent 6 }}\n  template:\n    metadata:\n      labels:"},
		{Range: protocol.Range{Start: protocol.Position{Line: 8, Character: 13}, End: protocol.Position{Line: 9, Character: 0}}, Text: "\n      \n"},
		{Range: protocol.Range{Start: protocol.Position{Line: 9, Character: 6}, End: protocol.Position{Line: 9, Character: 0}}, Text: "{{- if .Values.serviceAccount.create -}}\napiVersion: v1\nkind: ServiceAccount\nmetadata:\n  name: {{ include \"hello-world.serviceAccountName\" . }}\n  labels:\n    {{- include \"hello-world.labels\" . | nindent 4 }}\n  {{- with .Values.serviceAccount.annotations }}\n  annotations:\n    {{- toYaml . | nindent 4 }}\n  {{- end }}\n{{- end }}"},
		{Range: protocol.Range{Start: protocol.Position{Line: 17, Character: 0}, End: protocol.Position{Line: 17, Character: 0}}, Text: ""},
		{Range: protocol.Range{Start: protocol.Position{Line: 18, Character: 0}, End: protocol.Position{Line: 19, Character: 0}}, Text: ""},
	})

	print(doc.Content)
	expected := `{{ .Values.test }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "hello-world.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
      {{- if .Values.serviceAccount.create -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "hello-world.serviceAccountName" . }}
  labels:
    {{- include "hello-world.labels" . | nindent 4 }}
  {{- with .Values.serviceAccount.annotations }}
  annotations:
  {{- end }}
{{- end }}
`

	assert.Equal(expected, doc.Content)
}
