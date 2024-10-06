package handler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestLoadDocsOnNewChart(t *testing.T) {
	tempDir := t.TempDir()
	rootURI := uri.File(tempDir)

	templateDir := filepath.Join(tempDir, "templates")
	err := os.MkdirAll(templateDir, 0o755)
	assert.NoError(t, err)

	templateFiles := []string{
		filepath.Join(templateDir, "template1.txt"),
		filepath.Join(templateDir, "template2.txt"),
	}

	for _, file := range templateFiles {
		err = os.WriteFile(file, []byte("This is a template file"), 0o644)
		assert.NoError(t, err)
	}

	h := &ServerHandler{
		documents:    lsplocal.NewDocumentStore(),
		helmlsConfig: util.DefaultConfig,
	}

	h.LoadDocsOnNewChart(charts.NewChart(rootURI, util.DefaultConfig.ValuesFilesConfig))

	for _, file := range templateFiles {
		doc, ok := h.documents.GetTemplateDoc(uri.File(file))
		assert.True(t, ok)
		assert.NotNil(t, doc)
		assert.False(t, doc.IsOpen)
	}
}

func TestLoadDocsOnNewChartDoesNotOverwrite(t *testing.T) {
	tempDir := t.TempDir()
	rootURI := uri.File(tempDir)

	templateDir := filepath.Join(tempDir, "templates")
	err := os.MkdirAll(templateDir, 0o755)
	assert.NoError(t, err)

	templateFile := filepath.Join(templateDir, "template1.txt")

	err = os.WriteFile(templateFile, []byte("This is a template file"), 0o644)
	assert.NoError(t, err)

	docs := lsplocal.NewDocumentStore()
	h := &ServerHandler{
		documents:    docs,
		helmlsConfig: util.DefaultConfig,
	}

	docs.DidOpenTemplateDocument(&lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI: uri.File(templateFile),
		},
	}, util.DefaultConfig)

	h.LoadDocsOnNewChart(charts.NewChart(rootURI, util.DefaultConfig.ValuesFilesConfig))

	doc, ok := h.documents.GetTemplateDoc(uri.File(templateFile))
	assert.True(t, ok)
	assert.NotNil(t, doc)
	// The document should still be open because it's already in the store
	assert.True(t, doc.IsOpen)
}

func TestLoadDocsOnNewChartWorksForMissingTemplateDir(t *testing.T) {
	tempDir := t.TempDir()
	rootURI := uri.File(tempDir)

	docs := lsplocal.NewDocumentStore()
	h := &ServerHandler{
		documents:    docs,
		helmlsConfig: util.DefaultConfig,
	}

	h.LoadDocsOnNewChart(charts.NewChart(rootURI, util.DefaultConfig.ValuesFilesConfig))

	h.LoadDocsOnNewChart(charts.NewChart(uri.File("NonExisting"), util.DefaultConfig.ValuesFilesConfig))
}
