package templatehandler

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestCompletionMain(t *testing.T) {
	testCases := []struct {
		desc                   string
		position               lsp.Position
		expectedInsertText     string
		notExpectedInsertTexts []string
		expectedError          error
	}{
		{
			desc: "Test completion on {{ not }}",
			position: lsp.Position{
				Line:      11,
				Character: 7,
			},
			expectedInsertText: "toYaml",
			notExpectedInsertTexts: []string{
				"replicaCount",
			},
			expectedError: nil,
		},
		{
			desc: "Test completion on text",
			position: lsp.Position{
				Line:      4,
				Character: 0,
			},
			expectedInsertText: "{{- if $1 }}\n $2 \n{{- else }}\n $0 \n{{- end }}",
			notExpectedInsertTexts: []string{
				"replicaCount",
				"toYaml",
			},
			expectedError: nil,
		},
		{
			desc: "Test completion on {{ if (and .Values.    ) }}",
			position: lsp.Position{
				Line:      8,
				Character: 19,
			},
			expectedInsertText: "replicaCount",
			notExpectedInsertTexts: []string{
				helmdocs.HelmFuncs[0].Name,
			},
			expectedError: nil,
		},
		{
			desc: "Test completion on .Chart.N",
			position: lsp.Position{
				Line:      5,
				Character: 11,
			},
			expectedInsertText: "Name",
			notExpectedInsertTexts: []string{
				helmdocs.HelmFuncs[0].Name,
				"replicaCount",
				"toYaml",
			},
			expectedError: nil,
		},
		{
			desc: "Test completion on .Values.",
			position: lsp.Position{
				Line:      0,
				Character: 11,
			},
			expectedInsertText: "replicaCount",
			notExpectedInsertTexts: []string{
				helmdocs.HelmFuncs[0].Name,
			},
			expectedError: nil,
		},
		{
			desc: "Test completion on {{ . }}",
			position: lsp.Position{
				Line:      6,
				Character: 4,
			},
			expectedInsertText: "Chart",
			notExpectedInsertTexts: []string{
				helmdocs.HelmFuncs[0].Name,
				"js",
				"replicaCount",
				"toYaml",
			},
			expectedError: nil,
		},
		{
			desc: "Test completion on .Values.re",
			position: lsp.Position{
				Line:      1,
				Character: 13,
			},
			expectedInsertText: "replicaCount",
			notExpectedInsertTexts: []string{
				helmdocs.HelmFuncs[0].Name,
			},
			expectedError: nil,
		},
		{
			desc: "Test completion on {{ toY }}",
			position: lsp.Position{
				Line:      3,
				Character: 6,
			},
			expectedInsertText: "toYaml",
			notExpectedInsertTexts: []string{
				"replicaCount",
			},
			expectedError: nil,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			path := "../../../testdata/example/templates/completion-test.yaml"
			fileURI := uri.File(path)

			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal("Could not read test file", err)
			}
			result, err := completionTestCall(fileURI, string(content), tt.position)
			assert.Equal(t, tt.expectedError, err)
			assert.NotNil(t, result)

			if result == nil {
				return
			}

			insertTexts := []string{}
			for _, item := range result.Items {
				insertTexts = append(insertTexts, item.InsertText)
			}
			assert.Contains(t, insertTexts, tt.expectedInsertText)

			for _, notExpectedInsertText := range tt.notExpectedInsertTexts {
				assert.NotContains(t, insertTexts, notExpectedInsertText)
			}
		})
	}
}

func TestCompletionMainSingleLines(t *testing.T) {
	testCases := []struct {
		templateWithMark       string
		expectedInsertTexts    []string
		notExpectedInsertTexts []string
		err                    error
	}{
		{`Test completion on {{ $variable := "hello" }} {{ $v^ }}`, []string{"$variable"}, []string{".Values", "include"}, nil},
		{`Test completion on {{ $variable := "hello" }} {{ $^ }}`, []string{"$variable"}, []string{".Values", "include"}, nil},
		{"Test completion on {{ .Values.^ingress.hosts }}", []string{"ingress"}, []string{"hosts"}, nil},
		{"Test completion on {{.Bad.^}}", []string{}, []string{}, errors.New("[Bad ] is no valid template context for helm")},
		{"Test completion on {{ .Bad.^ }}", []string{}, []string{}, errors.New("[Bad ] is no valid template context for helm")},
		{"Test completion on {{ n^ }}", []string{"not"}, []string{}, nil},
		{"Test completion on {{ .Values.^ }}", []string{"replicaCount"}, []string{}, nil},
		{"Test completion on {{ .Release.N^ }}", []string{"Name"}, []string{}, nil},
		{"Test completion on {{ .Capabilities.KubeVersion.^ }}", []string{"Minor"}, []string{}, nil},
		{"Test completion on {{ .Capabilities.KubeVersion.Mi^ }}", []string{"nor"}, []string{}, nil},
		{`Test completion on {{ define "test" }} T1 {{ end }} {{ include "te^"}}`, []string{"test"}, []string{}, nil},
		{`Test completion on {{ range .Values.ingress.hosts }} {{ .^ }} {{ end }}`, []string{"host", "paths"}, []string{}, nil},
		{`Test completion on {{ range .Values.ingress.hosts }} {{ .ho^  }} {{ end }}`, []string{"host", "paths"}, []string{}, nil},
		{`Test completion on {{ range .Values.ingress.hosts }} {{ range .paths 	}} {{ .^ }} {{ end }} {{ end }}`, []string{"pathType", "path"}, []string{}, nil},
		{`Test completion on {{ $root := . }} {{ $root.test.^ }}`, []string{}, []string{}, errors.New("[test ] is no valid template context for helm")},
		{`Test completion on {{ range $type, $config := $.Values.deployments }} {{ .^ }} {{ end }}`, []string{"some"}, []string{}, nil},
		{`Test completion on {{ range $type, $config := $.Values.deployments }} {{ .s^ }} {{ end }}`, []string{"some"}, []string{}, nil},
		{`Test completion on {{ range $type, $config := $.Values.deployments }} {{ $config.^ }} {{ end }}`, []string{"some"}, []string{}, nil},
		{`Test completion on {{ range .Values.deploymentsWithNestedStuff }} {{ .hpa.cpuUtilization.^ }} {{ end }}`, []string{"targetAverageUtilization", "enabled"}, []string{}, nil},
		{`Test completion on {{ range $type, $config := .Values.deploymentsWithNestedStuff }} {{ .hpa.cpuUtilization.^ }} {{ end }}`, []string{"targetAverageUtilization", "enabled"}, []string{}, nil},
		{`Test completion on {{ range $type, $config := .Values.deploymentsWithNestedStuff }} {{ $config.hpa.cpuUtilization.^ }} {{ end }}`, []string{"targetAverageUtilization", "enabled"}, []string{}, nil},
		{`Test completion on {{ range $type, $config := $.Values.deployments }} {{ $config.s^ }} {{ end }}`, []string{"some"}, []string{}, nil},
	}

	for _, tt := range testCases {
		t.Run(tt.templateWithMark, func(t *testing.T) {
			pos, buf := getPositionForMarkedTestLine(tt.templateWithMark)

			// to get the correct values file ../../../testdata/example/values.yaml
			fileURI := uri.File("../../../testdata/example/templates/completion-test.yaml")

			result, err := completionTestCall(fileURI, buf, pos)
			assert.NotNil(t, result)
			assert.Equal(t, tt.err, err)

			if result == nil {
				return
			}

			insertTexts := []string{}
			for _, item := range result.Items {
				insertTexts = append(insertTexts, item.InsertText)
			}

			for _, expectedInsertText := range tt.expectedInsertTexts {
				assert.Contains(t, insertTexts, expectedInsertText)
			}
			for _, notExpectedInsertText := range tt.notExpectedInsertTexts {
				assert.NotContains(t, insertTexts, notExpectedInsertText)
			}
		})
	}
}

func completionTestCall(fileURI uri.URI, buf string, pos lsp.Position) (*lsp.CompletionList, error) {
	documents := document.NewDocumentStore()
	d := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:        fileURI,
			LanguageID: "",
			Version:    0,
			Text:       buf,
		},
	}
	documents.DidOpenTemplateDocument(&d, util.DefaultConfig)
	h := &TemplateHandler{
		chartStore:      charts.NewChartStore(uri.File("."), charts.NewChart, func(chart *charts.Chart) {}),
		documents:       documents,
		yamllsConnector: &yamlls.Connector{},
	}
	result, err := h.Completion(context.Background(), &lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: fileURI,
			},
			Position: pos,
		},
	})
	return result, err
}
