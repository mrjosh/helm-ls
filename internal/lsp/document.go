package lsp

import (
	"bytes"
	"strings"

	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/pkg/errors"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// documentStore holds opened documents.
type DocumentStore struct {
	documents map[string]*Document
	fs        FileStorage
}

func NewDocumentStore(fs FileStorage) *DocumentStore {
	return &DocumentStore{
		documents: map[string]*Document{},
		fs:        fs,
	}
}

func (s *DocumentStore) GetAllDocs() []*Document {
	var docs []*Document
	for _, doc := range s.documents {
		docs = append(docs, doc)
	}
	return docs
}

func (s *DocumentStore) DidOpen(params lsp.DidOpenTextDocumentParams,helmlsConfig util.HelmlsConfiguration) (*Document, error) {
	//langID := params.TextDocument.LanguageID
	//if langID != "markdown" && langID != "vimwiki" && langID != "pandoc" {
		//return nil, nil
	//}

	uri := params.TextDocument.URI
	path, err := s.normalizePath(uri)
	if err != nil {
		return nil, err
	}
	doc := &Document{
		URI:     uri,
		Path:    path,
		Content: params.TextDocument.Text,
		Ast:  ParseAst(params.TextDocument.Text),
		DiagnosticsCache: NewDiagnosticsCache(helmlsConfig),
	}
	s.documents[path] = doc
	return doc, nil
}

func (s *DocumentStore) Close(uri lsp.DocumentURI) {
	delete(s.documents, uri.Filename())
}

func (s *DocumentStore) Get(docuri uri.URI) (*Document, bool) {
	path, err := s.normalizePath(docuri)
	if err != nil {
		logger.Debug(err)
		return nil, false
	}
	d, ok := s.documents[path]
	return d, ok
}

func (s *DocumentStore) normalizePath(docuri uri.URI) (string, error) {
	path, err := util.URIToPath(docuri)
	if err != nil {
		return "", errors.Wrapf(err, "unable to parse URI: %s", docuri)
	}
	return s.fs.Canonical(path), nil
}

// Document represents an opened file.
type Document struct {
	URI                     lsp.DocumentURI
	Path                    string
	NeedsRefreshDiagnostics bool
	Content                 string
	lines                   []string
	Ast                     *sitter.Tree
	DiagnosticsCache        diagnosticsCache        
}

// ApplyChanges updates the content of the document from LSP textDocument/didChange events.
func (d *Document) ApplyChanges(changes []lsp.TextDocumentContentChangeEvent) {
	var content = []byte(d.Content)
	for _, change := range changes {
		start, end := util.PositionToIndex(change.Range.Start, content), util.PositionToIndex(change.Range.End, content)

		var buf bytes.Buffer
		buf.Write(content[:start])
		buf.Write([]byte(change.Text))
		buf.Write(content[end:])
		content = buf.Bytes()
	}
	d.Content = string(content)

	d.ApplyChangesToAst(d.Content)

	d.lines = nil
}

// WordAt returns the word found at the given location.
func (d *Document) WordAt(pos lsp.Position) string {

  logger.Debug(pos)

	line, ok := d.GetLine(int(pos.Line))
	if !ok {
		return ""
	}
	return util.WordAt(line, int(pos.Character))
}

func (d *Document) ValueAt(pos lsp.Position) string {
  logger.Debug(pos)

	line, ok := d.GetLine(int(pos.Line))
	if !ok {
		return ""
	}
	return util.ValueAt(line, int(pos.Character))
}

// ContentAtRange returns the document text at given range.
func (d *Document) ContentAtRange(rng lsp.Range) string {
	return d.Content[rng.Start.Character:rng.End.Character]
}

// GetLine returns the line at the given index.
func (d *Document) GetLine(index int) (string, bool) {
	lines := d.GetLines()
	if index < 0 || index > len(lines) {
		return "", false
	}
	return lines[index], true
}

// GetLines returns all the lines in the document.
func (d *Document) GetLines() []string {
	if d.lines == nil {
		// We keep \r on purpose, to avoid messing up position conversions.
		d.lines = strings.Split(d.Content, "\n")
	}
	return d.lines
}

// LookBehind returns the n characters before the given position, on the same line.
func (d *Document) LookBehind(pos lsp.Position, length int) string {
	line, ok := d.GetLine(int(pos.Line))
	if !ok {
		return ""
	}

	charIdx := int(pos.Character)
	if length > charIdx {
		return line[0:charIdx]
	}
	return line[(charIdx - length):charIdx]
}

// LookForward returns the n characters after the given position, on the same line.
func (d *Document) LookForward(pos lsp.Position, length int) string {
	line, ok := d.GetLine(int(pos.Line))
	if !ok {
		return ""
	}

	lineLength := len(line)
	charIdx := int(pos.Character)
	if lineLength <= charIdx+length {
		return line[charIdx:]
	}
	return line[charIdx:(charIdx + length)]
}

// LinkFromRoot returns a Link to this document from the root of the given
// notebook.
//func (d *document) LinkFromRoot(nb *core.Notebook) (*documentLink, error) {
	//href, err := nb.RelPath(d.Path)
	//if err != nil {
		//return nil, err
	//}
	//return &documentLink{
		//Href:          href,
		//RelativeToDir: nb.Path,
	//}, nil
//}

// DocumentLinkAt returns the internal or external link found in the document
// at the given position.
//func (d *Document) DocumentLinkAt(pos lsp.Position) (*documentLink, error) {
	//links, err := d.DocumentLinks()
	//if err != nil {
		//return nil, err
	//}

	//for _, link := range links {
		//if positionInRange(d.Content, link.Range, pos) {
			//return &link, nil
		//}
	//}

	//return nil, nil
//}

// DocumentLinks returns all the internal and external links found in the
// document.
//func (d *document) DocumentLinks() ([]documentLink, error) {
	//links := []documentLink{}

	//lines := d.GetLines()
	//for lineIndex, line := range lines {

		//appendLink := func(href string, start, end int, hasTitle bool, isWikiLink bool) {
			//if href == "" {
				//return
			//}

			//// Go regexes work with bytes, but the LSP client expects character indexes.
			//start = strutil.ByteIndexToRuneIndex(line, start)
			//end = strutil.ByteIndexToRuneIndex(line, end)

			//links = append(links, documentLink{
				//Href:          href,
				//RelativeToDir: filepath.Dir(d.Path),
				//Range: protocol.Range{
					//Start: protocol.Position{
						//Line:      protocol.UInteger(lineIndex),
						//Character: protocol.UInteger(start),
					//},
					//End: protocol.Position{
						//Line:      protocol.UInteger(lineIndex),
						//Character: protocol.UInteger(end),
					//},
				//},
				//HasTitle:   hasTitle,
				//IsWikiLink: isWikiLink,
			//})
		//}

		//for _, match := range markdownLinkRegex.FindAllStringSubmatchIndex(line, -1) {
			//// Ignore embedded image, e.g. ![title](href.png)
			//if match[0] > 0 && line[match[0]-1] == '!' {
				//continue
			//}

			//href := line[match[4]:match[5]]
			//// Valid Markdown links are percent-encoded.
			//if decodedHref, err := url.PathUnescape(href); err == nil {
				//href = decodedHref
			//}
			//appendLink(href, match[0], match[1], false, false)
		//}

		//for _, match := range wikiLinkRegex.FindAllStringSubmatchIndex(line, -1) {
			//href := line[match[2]:match[3]]
			//hasTitle := match[4] != -1
			//appendLink(href, match[0], match[1], hasTitle, true)
		//}
	//}

	//return links, nil
//}

//// IsTagPosition returns whether the given caret position is inside a tag (YAML frontmatter, #hashtag, etc.).
//func (d *document) IsTagPosition(position protocol.Position, noteContentParser core.NoteContentParser) bool {
	//lines := strutil.CopyList(d.GetLines())
	//lineIdx := int(position.Line)
	//charIdx := int(position.Character)
	//line := lines[lineIdx]
	//// https://github.com/mickael-menu/zk/issues/144#issuecomment-1006108485
	//line = line[:charIdx] + "ZK_PLACEHOLDER" + line[charIdx:]
	//lines[lineIdx] = line
	//targetWord := strutil.WordAt(line, charIdx)
	//if targetWord == "" {
		//return false
	//}

	//content := strings.Join(lines, "\n")
	//note, err := noteContentParser.ParseNoteContent(content)
	//if err != nil {
		//return false
	//}
	//return strutil.Contains(note.Tags, targetWord)
//}

//type documentLink struct {
	//Href          string
	//RelativeToDir string
	//Range         lsp.Range
	//// HasTitle indicates whether this link has a title information. For
	//// example [[filename]] doesn't but [[filename|title]] does.
	//HasTitle bool
	//// IsWikiLink indicates whether this link is a [[WikiLink]] instead of a
	//// regular Markdown link.
	//IsWikiLink bool
//}
