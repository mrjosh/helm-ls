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

var addChartCallback = func(chart *charts.Chart) {}

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
			documents.DidOpenTemplateDocument(&d, util.DefaultConfig)
			h := &langHandler{
				chartStore:      charts.NewChartStore(uri.File("."), charts.NewChart, addChartCallback),
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
			documents.DidOpenTemplateDocument(&d, util.DefaultConfig)
			h := &langHandler{
				chartStore:      charts.NewChartStore(uri.File("."), charts.NewChart, addChartCallback),
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

func TestRefercesSingleLines(t *testing.T) {
	testCases := []struct {
		templateWithMark             string
		expectedReferencesStartChars []int
		err                          error
	}{
		{"Test References on {{ $test := .Values.value }} {{ $te^st }}", []int{22, 51}, nil},
		{"Test References on {{ range $test := .Values.value }} {{ $te^st }}", []int{28, 57}, nil},
		{"Test References on {{ range $test := .Values.value }} {{ $te^st }} {{ end }} {{ $test }}", []int{28, 57}, nil},
		{"Test References on {{ range $test := .Values.value }} {{ $te^st }}  {{ $test }} {{ end }}", []int{28, 57, 70}, nil},
		{"Test References on {{ if .Values.bla }} {{ $test := 2 }} {{ $t^est }} {{ end }} {{ $test := 3 }} {{ $test }}", []int{43, 60}, nil},
		{"Test References on {{ if .Values.bla }} {{ $test := 2 }} {{ $test }} {{ end }} {{ $te^st := 3 }} {{ $test }}", []int{82, 99}, nil},

		{`{{define "na^me"}} T1 {{end}} {{include "name" .}} {{include "name" .}} `, []int{0, 31, 52}, nil},
		{`{{define "name"}} T1 {{end}} {{include "na^me" .}} {{include "name" .}} `, []int{0, 31, 52}, nil},
		{`{{define "name"}} T1 {{end}} {{include "name" .}} {{include "nam^e" .}} `, []int{0, 31, 52}, nil},
	}

	for _, tt := range testCases {
		t.Run(tt.templateWithMark, func(t *testing.T) {
			documents := lsplocal.NewDocumentStore()
			pos, buf := getPositionForMarkedTestLine(tt.templateWithMark)
			fileURI := uri.File("fake-testfile.yaml")

			d := lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI:        fileURI,
					LanguageID: "",
					Version:    0,
					Text:       buf,
				},
			}
			documents.DidOpenTemplateDocument(&d, util.DefaultConfig)
			h := &langHandler{
				chartStore:      charts.NewChartStore(uri.File("."), charts.NewChart, addChartCallback),
				documents:       documents,
				yamllsConnector: &yamlls.Connector{},
			}
			result, err := h.References(context.Background(), &lsp.ReferenceParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{
						URI: fileURI,
					},
					Position: pos,
				},
			})
			assert.Equal(t, tt.err, err)
			startPointChars := []int{}
			for _, location := range result {
				startPointChars = append(startPointChars, int(location.Range.Start.Character))
			}
			assert.ElementsMatch(t, tt.expectedReferencesStartChars, startPointChars)
		})
	}
}
