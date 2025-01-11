//go:build integration

package yamlls

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
      "type": "string",
  		"description": "Post office box number"
    },
    "countryName": {
      "type": "string",
			"description": "Country name"
    }
  }
}
`

func TestYamllsCustomSchemaProviderDiagnosticsIntegration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Test does not work on windows for unknown reasons")
	}
	_, diagnosticsChan, _ := getYamllsConnectorWithCustomSchema(t)
	diagnostic := []lsp.Diagnostic{}
	afterCh := time.After(20 * time.Second)
	for {
		if len(diagnostic) > 0 {
			break
		}
		select {
		case d := <-diagnosticsChan:
			diagnostic = append(diagnostic, d.Diagnostics...)
		case <-afterCh:
			t.Fatal("Timed out waiting for diagnostics")
		}
	}

	assert.Len(t, diagnostic, 1)
	assert.Equal(t, "Yamlls: Incorrect type. Expected \"address\".", diagnostic[0].Message)
}

func TestYamllsCustomSchemaProviderCompletionIntegration(t *testing.T) {
	yamllsConnector, _, testFile := getYamllsConnectorWithCustomSchema(t)

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
		t.Log("Called completion")

		assert.NotNil(c, result)
		if result == nil {
			t.Log("result is nil")
			return
		}

		items := result.Items
		assert.Len(c, items, 2)

		assert.Equal(c, "postOfficeBox: ", items[0].InsertText)
		assert.Equal(c, "countryName: ", items[1].InsertText)
	}, time.Second*10, time.Second*2)
}

func TestYamllsCustomSchemaProviderHoverIntegration(t *testing.T) {
	yamllsConnector, _, testFile := getYamllsConnectorWithCustomSchema(t)

	yamllsConnector.DocumentDidChange(&lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			Version: 1,
			TextDocumentIdentifier: lsp.TextDocumentIdentifier{
				URI: uri.File(testFile),
			},
		},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{
			{
				Text: "postOfficeBox: ",
			},
		},
	})

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		result, _ := yamllsConnector.CallHover(context.Background(), lsp.HoverParams{
			TextDocumentPositionParams: lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: uri.File(testFile),
				},
				Position: lsp.Position{
					Line:      0,
					Character: 1,
				},
			},
		}, "postOfficeBox")
		t.Log("Called completion")

		assert.NotNil(c, result)
		if result == nil {
			t.Log("result is nil")
			return
		}

		assert.True(c, strings.HasPrefix(result.Contents.Value, "Post office box number"))
	}, time.Second*10, time.Second*2)
}

func getYamllsConnectorWithCustomSchema(t *testing.T) (*Connector, chan lsp.PublishDiagnosticsParams, string) {
	config := util.DefaultConfig.YamllsConfiguration
	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "schema.json")
	err := os.WriteFile(schemaFile, []byte(TEST_JSON_SCHEMA), 0o644)
	assert.NoError(t, err)

	testFile := filepath.Join(tempDir, "test.yaml")
	err = os.WriteFile(testFile, []byte("c"), 0o644)
	assert.NoError(t, err)

	customHandler := NewCustomSchemaHandler(
		NewCustomSchemaProviderHandler(
			func(ctx context.Context, URI uri.URI) (uri.URI, error) {
				log.Printf("Custom schema provider called for %s returning %s", URI, uri.File(schemaFile))
				return uri.File(schemaFile), nil
			}))

	yamllsConnector, documents, diagnosticsChan := getYamllsConnector(t, config, customHandler)

	openFile(t, documents, testFile, yamllsConnector)
	return yamllsConnector, diagnosticsChan, testFile
}
