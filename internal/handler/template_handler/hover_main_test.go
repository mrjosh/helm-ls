package templatehandler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
			desc: "Test hover on template context in range over mapping",
			position: lsp.Position{
				Line:      85,
				Character: 26,
			},
			expected:      fmt.Sprintf("### %s\n%s\n", filepath.Join("..", "..", "..", "testdata", "example", "values.yaml"), "```yaml\nvalue\n```"),
			expectedError: nil,
		},
		{
			desc: "Test hover on template context with variables",
			position: lsp.Position{
				Line:      74,
				Character: 50,
			},
			expected:      fmt.Sprintf("### %s\n%s\n", filepath.Join("..", "..", "..", "testdata", "example", "values.yaml"), "```yaml\nfirst:\n  some: value\nsecond:\n  some: value\n```"),
			expectedError: nil,
		},
		{
			desc: "Test hover on template context with variables in range loop",
			position: lsp.Position{
				Line:      80,
				Character: 31,
			},
			expected:      fmt.Sprintf("### %s\n%s\n", filepath.Join("..", "..", "..", "testdata", "example", "values.yaml"), "```yaml\nvalue\n```"),
			expectedError: nil,
		},
		{
			desc: "Test hover on dot",
			position: lsp.Position{
				Line:      17,
				Character: 19,
			},
			expected:      fmt.Sprintf("### %s\n%s\n", filepath.Join("..", "..", "..", "testdata", "example", "values.yaml"), "```yaml\n{}\n```"),
			expectedError: nil,
		},
		{
			desc: "Test hover on function",
			position: lsp.Position{
				Line:      7,
				Character: 10,
			},
			expected:      "not $x\n\nNegate the boolean value of $x",
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
			expected:      fmt.Sprintf("### %s\n%s\n", filepath.Join("..", "..", "..", "testdata", "example", "values.yaml"), "```yaml\nimagePullSecrets: []\n```"),
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
			expected:      "Access non-template files within the chart",
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
			expected:      fmt.Sprintf("### %s\n%s\n", filepath.Join("..", "..", "..", "testdata", "example", "values.yaml"), "```yaml\ningress.hosts:\n- host: chart-example.local\n  paths:\n  - path: /\n    pathType: ImplementationSpecific\n```"),
			expectedError: nil,
		},
		{
			desc: "Test hover values number",
			position: lsp.Position{
				Line:      8,
				Character: 28,
			},
			expected:      fmt.Sprintf("### %s\n%s\n", filepath.Join("..", "..", "..", "testdata", "example", "values.yaml"), "```yaml\n1\n```"),
			expectedError: nil,
		},
		{
			desc: "Test hover include parameter",
			position: lsp.Position{
				Line:      3,
				Character: 28,
			},
			expected: "### " + filepath.Join("..", "_helpers.tpl") + "\n```helm\n" + `{{- define "example.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}
` + "```\n",
			expectedError: nil,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			documents := lsplocal.NewDocumentStore()

			path := "../../../testdata/example/templates/deployment.yaml"
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

			chart_store := charts.NewChartStore(uri.File("."), charts.NewChart, addChartCallback)
			h := &TemplateHandler{
				chartStore:      chart_store,
				documents:       documents,
				yamllsConnector: &yamlls.Connector{},
			}
			h.chartStore = charts.NewChartStore(uri.File("."), charts.NewChart, func(chart *charts.Chart) {
				documents.LoadDocsOnNewChart(chart, util.DefaultConfig)
			})
			chart, _ := h.chartStore.GetChartOrParentForDoc(fileURI)
			documents.LoadDocsOnNewChart(chart, util.DefaultConfig)
			result, err := h.Hover(context.Background(), &lsp.HoverParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{
						URI: fileURI,
					},
					Position: tt.position,
				},
			})
			assert.Equal(t, tt.expectedError, err)
			if tt.expected == "" {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, strings.ReplaceAll(result.Contents.Value, "\r\n", "\n"))
			}
		})
	}
}
