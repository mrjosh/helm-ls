package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestHoverMain(t *testing.T) {
	testCases := []struct {
		desc          string
		position      lsp.Position
		expected      string
		expectedError error
	}{
		{
			desc: "Test hover on function",
			position: lsp.Position{
				Line:      7,
				Character: 10,
			},
			expected:      "not $x\n\nnegate the boolean value of $x",
			expectedError: nil,
		},
		{
			desc: "Test hover on .Values",
			position: lsp.Position{
				Line:      25,
				Character: 18,
			},
			expected:      "The values made available through values.yaml, --set and -f.",
			expectedError: nil,
		},
		{
			desc: "Test hover on empty array in .Values",
			position: lsp.Position{
				Line:      25,
				Character: 28,
			},
			expected:      fmt.Sprintf("### %s\n%s\n\n\n", filepath.Join("..", "..", "testdata", "example", "values.yaml"), "imagePullSecrets: []"),
			expectedError: nil,
		},
		{
			desc: "Test hover on .Chart metadata",
			position: lsp.Position{
				Line:      33,
				Character: 28,
			},
			expected:      "Name of the chart\n\nexample\n",
			expectedError: nil,
		},
		{
			desc: "Test hover on dot",
			position: lsp.Position{
				Line:      17,
				Character: 19,
			},
			expected:      fmt.Sprintf("### %s\n%s\n\n\n", filepath.Join("..", "..", "testdata", "example", "values.yaml"), "{}"),
			expectedError: nil,
		},
		{
			desc: "Test hover on .Files function",
			position: lsp.Position{
				Line:      68,
				Character: 24,
			},
			expected:      "Returns a list of files whose names match the given shell glob pattern.",
			expectedError: nil,
		},
		{
			desc: "Test hover on .Files",
			position: lsp.Position{
				Line:      68,
				Character: 20,
			},
			expected:      "access non-template files within the chart",
			expectedError: nil,
		},
		{
			desc: "Test hover on yaml text",
			position: lsp.Position{
				Line:      0,
				Character: 0,
			},
			expected:      "",
			expectedError: nil,
		},
		{
			desc: "Test hover values list",
			position: lsp.Position{
				Line:      71,
				Character: 35,
			},
			expected:      fmt.Sprintf("### %s\n%s\n\n", filepath.Join("..", "..", "testdata", "example", "values.yaml"), "ingress.hosts:\n- host: chart-example.local\n  paths:\n  - path: /\n    pathType: ImplementationSpecific\n"),
			expectedError: nil,
		},
		{
			desc: "Test hover values number",
			position: lsp.Position{
				Line:      8,
				Character: 28,
			},
			expected:      fmt.Sprintf("### %s\n%s\n\n", filepath.Join("..", "..", "testdata", "example", "values.yaml"), "1"),
			expectedError: nil,
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
			result, err := h.Hover(context.Background(), &lsp.HoverParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{
						URI: fileURI,
					},
					Position: tt.position,
				},
			})
			assert.Equal(t, tt.expectedError, err)
			if result == nil {
				t.Fatal("Result is nil")
			}
			assert.Equal(t, tt.expected, result.Contents.Value)
		})
	}
}
