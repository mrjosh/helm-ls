//go:build integration

package yamlls

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestYamllsDocumentDidChangeFullSyncTemplateIntegration(t *testing.T) {
	config := util.DefaultConfig.YamllsConfiguration
	yamllsConnector, documents, _ := getYamllsConnector(t, config, &DefaultCustomHandler)

	file := filepath.Join(t.TempDir(), "templates", "service.yaml")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	openFile(t, documents, file, yamllsConnector)

	fileURI := uri.File(file)
	doc, ok := documents.GetTemplateDoc(fileURI)
	if !ok {
		t.Fatal("Could not get document")
	}

	// remove the metadata block, then sync the whole document as done on save
	doc.ApplyChanges([]lsp.TextDocumentContentChangeEvent{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 2, Character: 0},
				End:   lsp.Position{Line: 4, Character: 0},
			},
			Text: "",
		},
	})
	yamllsConnector.DocumentDidChangeFullSyncTemplate(doc, lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: lsp.TextDocumentIdentifier{URI: fileURI},
			Version:                1,
		},
	})

	// yamlls only suggests metadata if its copy no longer contains the removed block
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		result, _ := yamllsConnector.CallCompletion(context.Background(), &lsp.CompletionParams{
			TextDocumentPositionParams: lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: fileURI},
				Position:     lsp.Position{Line: 2, Character: 0},
			},
		})
		assert.NotNil(c, result)
		if result == nil {
			return
		}
		labels := []string{}
		for _, item := range result.Items {
			labels = append(labels, item.Label)
		}
		assert.Contains(c, labels, "metadata")
	}, time.Second*10, time.Second/2)
}
