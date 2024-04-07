package handler

import (
	"context"
	"os"
	"testing"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestRefercesIncludes(t *testing.T) {
	content := `{{define "name"}} T1 {{end}}
 {{include "name" .}}
 {{include "name" .}}
`

	expected := []lsp.Location{
		{
			URI: uri.File("/tmp/testfile.yaml"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line: 0x1, Character: 0x3,
				},
				End: protocol.Position{
					Line:      0x1,
					Character: 0x13,
				},
			},
		},
		protocol.Location{
			URI: uri.File("/tmp/testfile.yaml"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      0x2,
					Character: 0x3,
				},
				End: protocol.Position{
					Line:      0x2,
					Character: 0x13,
				},
			},
		},
		protocol.Location{
			URI: uri.File("/tmp/testfile.yaml"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      0x0,
					Character: 0x0,
				},
				End: protocol.Position{
					Line:      0x0,
					Character: 0x1c,
				},
			},
		},
	}
	testCases := []struct {
		desc          string
		position      lsp.Position
		expected      []lsp.Location
		expectedError error
	}{
		{
			desc: "Test references on define",
			position: lsp.Position{
				Line:      0,
				Character: 11,
			},
			expected: expected,
		},
		{
			desc: "Test references on include",
			position: lsp.Position{
				Line:      2,
				Character: 14,
			},
			expected: expected,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			documents := lsplocal.NewDocumentStore()

			path := "/tmp/testfile.yaml"
			fileURI := uri.File(path)

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
			result, err := h.References(context.Background(), &lsp.ReferenceParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{
						URI: fileURI,
					},
					Position: tt.position,
				},
			})
			assert.Equal(t, tt.expectedError, err)
			if err == nil {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestRefercesTemplateContext(t *testing.T) {
	content := `
{{ .Values.test }}
{{ .Values.test.nested }}
{{ .Values.test }}
`
	expectedValues := []lsp.Location{
		{
			URI: uri.File("/tmp/testfile.yaml"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line: 0x1, Character: 4,
				},
				End: protocol.Position{
					Line:      0x1,
					Character: 0xa,
				},
			},
		},
		protocol.Location{
			URI: uri.File("/tmp/testfile.yaml"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      0x2,
					Character: 0x4,
				},
				End: protocol.Position{
					Line:      0x2,
					Character: 0xa,
				},
			},
		},
		protocol.Location{
			URI: uri.File("/tmp/testfile.yaml"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      0x3,
					Character: 0x4,
				},
				End: protocol.Position{
					Line:      0x3,
					Character: 0xa,
				},
			},
		},
	}
	testCases := []struct {
		desc          string
		position      lsp.Position
		expected      []lsp.Location
		expectedError error
	}{
		{
			desc: "Test references on .Values",
			position: lsp.Position{
				Line:      1,
				Character: 8,
			},
			expected: expectedValues,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			documents := lsplocal.NewDocumentStore()

			path := "/tmp/testfile.yaml"
			fileURI := uri.File(path)

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
			result, err := h.References(context.Background(), &lsp.ReferenceParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{
						URI: fileURI,
					},
					Position: tt.position,
				},
			})
			assert.Equal(t, tt.expectedError, err)
			if err == nil {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestRefercesTemplateContextWithTestFile(t *testing.T) {
	testCases := []struct {
		desc                string
		position            lsp.Position
		expectedStartPoints []lsp.Position
		expectedError       error
	}{
		{
			desc: "Test references on .Values",
			position: lsp.Position{
				Line:      25,
				Character: 18,
			},
			expectedStartPoints: []lsp.Position{{Line: 25, Character: 16}, {Line: 31, Character: 20}},
		},
		{
			desc: "Test references on .Values.imagePullSecrets",
			position: lsp.Position{
				Line:      25,
				Character: 31,
			},
			expectedStartPoints: []lsp.Position{{Line: 25, Character: 23}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			documents := lsplocal.NewDocumentStore()

			path := "../../testdata/example/templates/deployment.yaml"
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
			result, err := h.References(context.Background(), &lsp.ReferenceParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{
						URI: fileURI,
					},
					Position: tt.position,
				},
			})
			assert.Equal(t, tt.expectedError, err)
			startPoints := []lsp.Position{}
			for _, location := range result {
				startPoints = append(startPoints, location.Range.Start)
			}
			for _, expected := range tt.expectedStartPoints {
				assert.Contains(t, startPoints, expected)
			}
		})
	}
}
