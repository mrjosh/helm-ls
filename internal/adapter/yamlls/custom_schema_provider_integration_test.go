//go:build integration

package yamlls

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

var TEST_JSON_SCHEMA = `
{
  "$id": "https://example.com/address.schema.json",
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "description": "An address similar to http://microformats.org/wiki/h-card",
  "type": "object",
  "properties": {
    "postOfficeBox": {
      "type": "string"
    },
    "countryName": {
      "type": "string"
    }
  }
}
`

func TestYamllsCustomSchemaProviderIntegration(t *testing.T) {
	config := util.DefaultConfig.YamllsConfiguration
	config.Path = "yamlls-debug.sh"

	// tempDir := t.TempDir()
	tempDir := "/data/data/com.termux/files/usr/tmp/"
	schemaFile := path.Join(tempDir, "schema.json")
	// write schema
	err := os.WriteFile(schemaFile, []byte(TEST_JSON_SCHEMA), 0o644)
	assert.NoError(t, err)

	testFile := path.Join(tempDir, "test.yaml")
	err = os.WriteFile(testFile, []byte("c"), 0o644)
	assert.NoError(t, err)

	customHandler := NewCustomSchemaHandler(
		NewCustomSchemaProviderHandler(
			func(ctx context.Context, URI uri.URI) (uri.URI, error) {
				t.Log("Calling Schema provider")
				return "http://localhost:8000/schema.json", nil
			}))

	yamllsConnector, documents, _ := getYamlLsConnector(t, config, customHandler)
	openFile(t, documents, testFile, yamllsConnector)

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		result, _ := yamllsConnector.CallCompletion(context.Background(), &lsp.CompletionParams{
			TextDocumentPositionParams: lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: uri.File(testFile),
				},
				Position: lsp.Position{
					Line:      0,
					Character: 1,
				},
			},
		})
		assert.NotNil(c, result)
		if result == nil {
			t.Log("result is nil")
			return
		}
		t.Log("reuslt  is", result)
	}, time.Second*20, time.Second*2)
}
