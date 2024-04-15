package handler

import (
	"context"
	"os"
	"testing"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
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
		// {
		// 	desc: "Test completion on {{ }}",
		// 	position: lsp.Position{
		// 		Line:      4,
		// 		Character: 3,
		// 	},
		// 	expectedInsertText: "toYaml",
		// 	notExpectedInsertTexts: []string{
		// 		"replicaCount",
		// 	},
		// 	expectedError: nil,
		// },
	}
	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			documents := lsplocal.NewDocumentStore()

			path := "../../testdata/example/templates/completion-test.yaml"
			fileURI := uri.File(path)

			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal("Could not read test file", err)
			}
			d := lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI:        fileURI,
					LanguageID: "",
					Version:    0,
					Text:       string(content),
				},
			}
			documents.DidOpen(&d, util.DefaultConfig)
			h := &langHandler{
				chartStore:      charts.NewChartStore(uri.File("."), charts.NewChart),
				documents:       documents,
				yamllsConnector: &yamlls.Connector{},
			}
			result, err := h.Completion(context.Background(), &lsp.CompletionParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{
						URI: fileURI,
					},
					Position: tt.position,
				},
			})
			assert.Equal(t, tt.expectedError, err)
			assert.NotNil(t, result)

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
