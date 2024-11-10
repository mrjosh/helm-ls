package document

import (
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v3"
)

// TemplateDocument represents an helm template file.
type YamlDocument struct {
	Document
	Node       yaml.Node
	ParsedYaml map[string]interface{}
	ParseErr   error
}

func (d *YamlDocument) GetDocumentType() DocumentType {
	return YamlDocumentType
}

func NewYamlDocument(fileURI uri.URI, content []byte, isOpen bool, helmlsConfig util.HelmlsConfiguration) *YamlDocument {
	node, err := util.ReadYamlToNode(content)

	logger.Debug("Parsing yaml", fileURI.Filename(), string(content), node.Value, err)

	var parsedYaml map[string]interface{}

	err = yaml.Unmarshal(content, &parsedYaml)
	return &YamlDocument{
		Document:   *NewDocument(fileURI, content, isOpen),
		Node:       node,
		ParsedYaml: parsedYaml,
		ParseErr:   err,
	}
}

// ApplyChanges updates the content of the document from LSP textDocument/didChange events.
func (d *YamlDocument) ApplyChanges(changes []lsp.TextDocumentContentChangeEvent) {
	d.Document.ApplyChanges(changes)

	node, err := util.ReadYamlToNode(d.Content)

	var parsedYaml map[string]interface{}
	err = yaml.Unmarshal(d.Content, &parsedYaml)

	d.Node = node
	d.ParsedYaml = parsedYaml
	d.ParseErr = err
}
