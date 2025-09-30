package document

import (
	"testing"

	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestNewYamlDocument(t *testing.T) {
	doc := NewYamlDocument(uri.File("test"), []byte("test: value"), true, util.DefaultConfig)
	assert.Equal(t, uri.File("test"), doc.URI)
	assert.Equal(t, []byte("test: value"), doc.Content)
	assert.Equal(t, true, doc.IsOpen)

	brokenYaml := `
test: fdsf

broken 
	`
	doc = NewYamlDocument(uri.File("test"), []byte(brokenYaml), true, util.DefaultConfig)
	assert.Error(t, doc.ParseErr)
	path, err := doc.GetPathForPosition(lsp.Position{Line: 0, Character: 0})
	assert.Error(t, err)
	assert.Equal(t, "", path)
}
