package yamlls

import (
	"encoding/json"
	"testing"
	"time"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

type readWriteCloseMock struct{}

func (proc readWriteCloseMock) Read(p []byte) (int, error) {
	return 1, nil
}

func (proc readWriteCloseMock) Write(p []byte) (int, error) {
	// TODO: remove the Content-Length header
	var params lsp.PublishDiagnosticsParams
	if err := json.Unmarshal(p, &params); err != nil {
		logger.Println("Error handling diagnostic", err)
	}

	logger.Printf("Write: %s", params)
	return 1, nil
}

func (proc readWriteCloseMock) Close() error {
	return nil
}

func TestYamllsDiagnosticsIntegration(t *testing.T) {
	dir := t.TempDir()
	documents := lsplocal.NewDocumentStore()
	con := jsonrpc2.NewConn(jsonrpc2.NewStream(readWriteCloseMock{}))
	config := util.DefaultConfig.YamllsConfiguration
	config.Path = "yamlls-debug.sh"
	yamllsConnector := NewConnector(config, con, documents)

	yamllsConnector.CallInitialize(uri.File(dir))

	content := `
{{- if .Values.serviceAccount.create }}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "whereabouts.serviceAccountName" . }}
  namespace: {{ include "common.names.namespace" . | quote }}
  labels: {{- include "common.labels.standard" ( dict "customLabels" .Values.commonLabels "context" $ ) | nindent 4 }}
  {{- if or .Values.serviceAccount.annotations .Values.commonAnnotations }}
  {{- $annotations := include "common.tplvalues.merge" ( dict "values" ( list .Values.serviceAccount.annotations .Values.commonAnnotations ) "context" . ) }}
  annotations: {{- include "common.tplvalues.render" ( dict "value" $annotations "context" $) | nindent 4 }}
  {{- end }}
automountServiceAccountToken: {{ .Values.serviceAccount.automountServiceAccountToken }}
{{- end }}
	`

	tree := lsplocal.ParseAst(nil, content)

	didOpenParams := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:        uri.File(dir + "/test.yaml"),
			LanguageID: "",
			Version:    0,
			Text:       content,
		},
	}
	documents.DidOpen(didOpenParams, util.DefaultConfig)
	yamllsConnector.DocumentDidOpen(tree, didOpenParams)

	logger.Printf("Running tests in %s", dir)

	// sleep 5 seconds
	time.Sleep(5 * time.Second)
}
