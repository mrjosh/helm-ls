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
	ParsedYaml map[string]any
	ParseErr   error
}

func (d *YamlDocument) GetDocumentType() DocumentType {
	return YamlDocumentType
}

func NewYamlDocument(fileURI uri.URI, content []byte, isOpen bool, helmlsConfig util.HelmlsConfiguration) *YamlDocument {
	node, parsedYaml, unmarshalErr := parseYaml(content)
	logger.Debug("Parsing yaml", fileURI.Filename(), string(content), node.Value, unmarshalErr)
	return &YamlDocument{
		Document:   *NewDocument(fileURI, content, isOpen),
		Node:       node,
		ParsedYaml: parsedYaml,
		ParseErr:   unmarshalErr,
	}
}

// ApplyChanges updates the content of the document from LSP textDocument/didChange events.
func (d *YamlDocument) ApplyChanges(changes []lsp.TextDocumentContentChangeEvent) {
	d.Document.ApplyChanges(changes)

	node, parsedYaml, unmarshalErr := parseYaml(d.Content)

	d.Node = node
	d.ParsedYaml = parsedYaml
	d.ParseErr = unmarshalErr
}

func parseYaml(content []byte) (yaml.Node, map[string]any, error) {
	node, err := util.ReadYamlToNode(content)
	if err != nil {
		return yaml.Node{}, map[string]any{}, err
	}

	var parsedYaml map[string]any
	unmarshalErr := yaml.Unmarshal(content, &parsedYaml)
	if unmarshalErr != nil {
		return node, map[string]any{}, unmarshalErr
	}

	return node, parsedYaml, nil
}
