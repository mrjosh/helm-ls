package util

import (
	"bytes"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
	"go.lsp.dev/protocol"
	lsp "go.lsp.dev/protocol"
)

func GetNodeForPosition(node ast.Node, position protocol.Position) ast.Node {
	visitor := &PositionFinderVisitor{
		position: position,
	}
	ast.Walk(visitor, node)
	return visitor.result
}

// PositionFinderVisitor is a visitor that collects positions.
type PositionFinderVisitor struct {
	position protocol.Position
	result   ast.Node
	found    bool
}

func (pv *PositionFinderVisitor) Visit(node ast.Node) ast.Visitor {
	if pv.found {
		return nil
	}

	if IsRelevantNode(node) && IsNodeAtPosition(node, &pv.position) {
		pv.found = true
		pv.result = node
		return nil
	}
	return pv
}

// We only care about user facing nodes
func IsRelevantNode(node ast.Node) bool {
	switch node.Type() {
	case ast.MappingKeyType, ast.MappingValueType, ast.MappingType, ast.SequenceType:
		return false
	default:
		return true
	}
}

func IsNodeAtPosition(node ast.Node, position *protocol.Position) bool {
	token := node.GetToken()
	if token == nil {
		return false
	}
	start := token.Position

	if start.Line != int(position.Line)+1 {
		return false
	}

	var endColumn int

	if token.Next != nil && token.Next.Position.Line == token.Position.Line {
		endColumn = token.Next.Position.Column - 1
	} else {
		endColumn = start.Column + len(node.String())
	}

	if start.Column <= int(position.Character)+1 && endColumn >= int(position.Character)+1 {
		return true
	}

	return false
}

// ReadYamlToNode will parse a YAML file into a yaml Node.
func ReadYamlToGoccyNode(data []byte) (node ast.Node, err error) {
	// --- WORKAROUND ---
	// Normalize Windows-style line endings (\r\n) to Unix-style (\n).
	// This prevents the tokenizer bug from miscalculating line numbers.
	// https://github.com/goccy/go-yaml/issues/560
	normalizedData := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))

	err = yaml.Unmarshal(normalizedData, &node)
	return node, err
}

func TokenToRange(token *token.Token) lsp.Range {
	if token == nil {
		return lsp.Range{}
	}

	return lsp.Range{
		Start: lsp.Position{Line: uint32(token.Position.Line - 1), Character: uint32(token.Position.Column - 1)},
		End:   lsp.Position{Line: uint32(token.Position.Line - 1), Character: uint32(token.Position.Column+len(token.Value)) + 1},
	}
}
