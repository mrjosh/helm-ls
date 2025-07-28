package document

import (
	"fmt"

	"github.com/goccy/go-yaml/ast"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v3"
)

// TemplateDocument represents an helm template file.
type YamlDocument struct {
	Document
	Node          yaml.Node
	GoccyYamlNode ast.Node
	ParsedYaml    map[string]any
	ParseErr      error
}

func (d *YamlDocument) GetDocumentType() DocumentType {
	return YamlDocumentType
}

func NewYamlDocument(fileURI uri.URI, content []byte, isOpen bool, helmlsConfig util.HelmlsConfiguration) *YamlDocument {
	node, goccyNode, parsedYaml, unmarshalErr := parseYaml(content)
	logger.Debug("Parsed yaml", fileURI.Filename(), unmarshalErr)
	return &YamlDocument{
		Document:      *NewDocument(fileURI, content, isOpen),
		Node:          node,
		GoccyYamlNode: goccyNode,
		ParsedYaml:    parsedYaml,
		ParseErr:      unmarshalErr,
	}
}

// ApplyChanges updates the content of the document from LSP textDocument/didChange events.
func (d *YamlDocument) ApplyChanges(changes []lsp.TextDocumentContentChangeEvent) {
	d.Document.ApplyChanges(changes)

	node, goccyNode, parsedYaml, unmarshalErr := parseYaml(d.Content)

	d.Node = node
	d.GoccyYamlNode = goccyNode
	d.ParsedYaml = parsedYaml
	d.ParseErr = unmarshalErr
}

func parseYaml(content []byte) (yaml.Node, ast.Node, map[string]any, error) {
	node, err := util.ReadYamlToNode(content)
	if err != nil {
		return yaml.Node{}, &ast.NullNode{}, map[string]any{}, err
	}

	goccyNode, err := util.ReadYamlToGoccyNode(content)
	if err != nil {
		return yaml.Node{}, &ast.NullNode{}, map[string]any{}, err
	}

	var parsedYaml map[string]any
	unmarshalErr := yaml.Unmarshal(content, &parsedYaml)
	if unmarshalErr != nil {
		return node, &ast.NullNode{}, map[string]any{}, unmarshalErr
	}

	return node, goccyNode, parsedYaml, nil
}

func (d *YamlDocument) GetPathForPosition(position lsp.Position) (string, error) {
	node := util.GetNodeForPosition(d.GoccyYamlNode, position)

	if node == nil {
		return "", fmt.Errorf("YAML node not found for position %v", position)
	}

	return node.GetPath(), nil
}
