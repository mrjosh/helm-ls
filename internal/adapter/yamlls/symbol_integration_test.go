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

func TestYamllsDocumentSymoblIntegration(t *testing.T) {
	config := util.DefaultConfig.YamllsConfiguration

	testCases := []struct {
		file        string
		expectedLen int
	}{
		{
			file:        "../../../testdata/example/templates/deployment.yaml",
			expectedLen: 4,
		},
		{
			file:        "../../../testdata/example/templates/ingress.yaml",
			expectedLen: 6,
		},
	}
	for _, tt1 := range testCases {
		tt := tt1
		t.Run(tt.file, func(t *testing.T) {
			t.Parallel()
			yamllsConnector, documents, _ := getYamlLsConnector(t, config)
			openFile(t, documents, tt.file, yamllsConnector)

			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				result, err := yamllsConnector.CallDocumentSymbol(context.Background(), &lsp.DocumentSymbolParams{
					TextDocument: lsp.TextDocumentIdentifier{
						URI: uri.File(tt.file),
					},
				})
				assert.NoError(c, err)
				assert.Len(c, result, tt.expectedLen)
			}, time.Second*10, time.Second*2)
		})
	}
}
