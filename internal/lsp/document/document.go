package document

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gobwas/glob"
	"github.com/mrjosh/helm-ls/internal/log"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

var logger = log.GetLogger()

// Document represents an opened file.
type DocumentType string

const (
	TemplateDocumentType DocumentType = "helm"
	YamlDocumentType     DocumentType = "yaml"
)

var ValuesYamlFileGlob = glob.MustCompile(`**/values*.{yaml,yml}`)

// Since not all editors support custom filetypes for values.yaml its easier to
// check if the file is a values.yaml or not on the server side
func DocumentTypeForFile(langID lsp.LanguageIdentifier, fileURI uri.URI) DocumentType {
	if strings.Contains(string(langID), `yaml`) {
		return YamlDocumentType
	}

	if IsValuesYamlFile(fileURI) {
		return YamlDocumentType
	}
	if strings.Contains(string(langID), `helm`) {
		return TemplateDocumentType
	}
	return TemplateDocumentType
}

func IsValuesYamlFile(fileURI uri.URI) bool {
	normalizedPath := strings.ReplaceAll(fileURI.Filename(), "\\", "/")
	return ValuesYamlFileGlob.Match(normalizedPath) && !strings.Contains(normalizedPath, "/templates/")
}

type Document struct {
	URI     lsp.DocumentURI
	Path    string
	Content []byte
	lines   []string
	IsOpen  bool
}

type DocumentInterface interface {
	GetDocumentType() DocumentType
	ApplyChanges([]lsp.TextDocumentContentChangeEvent)
}

func NewDocument(fileURI uri.URI, content []byte, isOpen bool) *Document {
	return &Document{
		URI:     fileURI,
		Path:    fileURI.Filename(),
		Content: content,
		IsOpen:  isOpen,
	}
}

// ApplyChanges updates the content of the document from LSP textDocument/didChange events.
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

// getLines returns all the lines in the document.
func (d *Document) getLines() []string {
	if d.lines == nil {
		// We keep \r on purpose, to avoid messing up position conversions.
		d.lines = strings.Split(string(d.Content), "\n")
	}
	return d.lines
}

// GetContent implements PossibleDependencyFile.
func (d *Document) GetContent() []byte {
	return d.Content
}

// GetPath implements PossibleDependencyFile.
func (d *Document) GetPath() string {
	return d.Path
}

type TextDocument interface {
	ApplyChanges([]lsp.TextDocumentContentChangeEvent)
}

// WordAt returns the word found at the given location.
func (d *Document) WordAt(pos lsp.Position) string {
	logger.Debug(pos)

	line, ok := d.getLine(int(pos.Line))
	if !ok {
		return ""
	}
	return util.WordAt(line, int(pos.Character))
}

// getLine returns the line at the given index.
func (d *Document) getLine(index int) (string, bool) {
	lines := d.getLines()
	if index < 0 || index > len(lines) {
		return "", false
	}
	return lines[index], true
}
