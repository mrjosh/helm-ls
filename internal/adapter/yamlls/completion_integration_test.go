//go:build integration

package yamlls

import (
	"context"
	"testing"
	"time"

	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestYamllsCompletionIntegration(t *testing.T) {
	config := util.DefaultConfig.YamllsConfiguration

	testCases := []struct {
		desc     string
		file     string
		position lsp.Position

		expected []lsp.CompletionItem
	}{
		{
			desc: "test completion on deployment.yaml",
			file: "../../../testdata/example/templates/deployment.yaml",
			position: lsp.Position{
				Line:      42,
				Character: 14,
			},
			expected: []lsp.CompletionItem{
				{
					Command:          nil,
					CommitCharacters: nil,
					Tags:             nil,
					Data:             nil,
					Deprecated:       false,
					Documentation:    "What host IP to bind the external port to.",
					FilterText:       "",
					InsertText:       "hostIP: ",
					InsertTextFormat: 2,
					Kind:             10,
					Label:            "hostIP",
					SortText:         "",
					TextEdit:         nil,
				},
				{
					Command:          nil,
					CommitCharacters: nil,
					Tags:             nil,
					Data:             nil,
					Deprecated:       false,
					Documentation:    "Number of port to expose on the host. If specified, this must be a valid port number, 0 < x < 65536. If HostNetwork is specified, this must match ContainerPort. Most containers do not need this.",
					FilterText:       "",
					InsertText:       "hostPort: ${1:0}",
					InsertTextFormat: 2,
					InsertTextMode:   0,
					Kind:             10,
					Label:            "hostPort",
					TextEdit:         nil,
				},
			},
		},
	}
	for _, tt1 := range testCases {
		tt := tt1
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()
			yamllsConnector, documents, _ := getYamlLsConnector(t, config)
			openFile(t, documents, tt.file, yamllsConnector)

			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				result, _ := yamllsConnector.CallCompletion(context.Background(), &lsp.CompletionParams{
					TextDocumentPositionParams: lsp.TextDocumentPositionParams{
						TextDocument: lsp.TextDocumentIdentifier{
							URI: uri.File(tt.file),
						},
						Position: tt.position,
					},
				})
				assert.NotNil(c, result)
				if result == nil {
					return
				}
				for i := 0; i < len(result.Items); i++ {
					result.Items[i].TextEdit = nil
				}
				for _, v := range tt.expected {
					assert.Contains(c, result.Items, v)
				}
			}, time.Second*10, time.Second/2)
		})
	}
}
