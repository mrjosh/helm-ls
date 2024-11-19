//go:build integration

package yamlls

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestYamllsHoverIntegration(t *testing.T) {
	config := util.DefaultConfig.YamllsConfiguration

	testCases := []struct {
		desc     string
		file     string
		position lsp.Position

		word     string
		expected string
	}{
		{
			desc: "test hover on deployment.yaml",
			file: "../../../testdata/example/templates/deployment.yaml",
			position: lsp.Position{
				Line:      0,
				Character: 0,
			},
			word:     "apiVersion",
			expected: "APIVersion defines the versioned schema of this representation of an object",
		},
		{
			desc: "test hover on deployment.yaml, template",
			file: "../../../testdata/example/templates/deployment.yaml",
			position: lsp.Position{
				Line:      13,
				Character: 3,
			},
			word:     "template",
			expected: "Template describes the pods that will be created",
		},
		{
			desc: "test hover on file without templates",
			file: "../../../testdata/example/templates/deployment-no-templates.yaml",
			position: lsp.Position{
				Line:      0,
				Character: 1,
			},
			word:     "",
			expected: "APIVersion defines the versioned schema of this representation of an object",
		},
	}
	for _, tt1 := range testCases {
		tt := tt1
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()
			yamllsConnector, documents, _ := getYamlLsConnector(t, config, &DefaultCustomHandler)
			openFile(t, documents, tt.file, yamllsConnector)

			assert.Eventually(t, func() bool {
				result, err := yamllsConnector.CallHover(context.Background(), lsp.HoverParams{
					TextDocumentPositionParams: lsp.TextDocumentPositionParams{
						TextDocument: lsp.TextDocumentIdentifier{
							URI: uri.File(tt.file),
						},
						Position: tt.position,
					},
				}, tt.word)
				return err == nil && strings.Contains(result.Contents.Value, tt.expected)
			}, time.Second*40, time.Second*2)
		})
	}
}
