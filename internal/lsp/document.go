package lsp

import (
	"bytes"
	"fmt"

	"github.com/mrjosh/helm-ls/internal/log"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

var logger = log.GetLogger()

// Document represents an opened file.
type Document struct {
	URI     lsp.DocumentURI
	Path    string
	Content []byte
	lines   []string
	IsOpen  bool
}

func (d *Document) ApplyChanges(changes []lsp.TextDocumentContentChangeEvent) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Sprintf("Recovered in ApplyChanges for %s, the document may be corrupted ", d.URI), r)
		}
	}()

	content := d.Content
	for _, change := range changes {
		start, end := util.PositionToIndex(change.Range.Start, content), util.PositionToIndex(change.Range.End, content)

		var buf bytes.Buffer
		buf.Write(content[:start])
		buf.Write([]byte(change.Text))
		buf.Write(content[end:])
		content = buf.Bytes()
	}
	d.Content = content
	d.lines = nil
}
