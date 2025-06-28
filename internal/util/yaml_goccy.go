package util

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"go.lsp.dev/protocol"
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
	start := token.Position

	endColumn := start.Column + len(node.String())

	if start.Line != int(position.Line)+1 {
		return false
	}

	if start.Column <= int(position.Character)+1 && endColumn >= int(position.Character)+1 {
		return true
	}
	return false
}

// ReadYamlToNode will parse a YAML file into a yaml Node.
func ReadYamlToGoccyNode(data []byte) (node ast.Node, err error) {
	err = yaml.Unmarshal(data, &node)
	return node, err
}
