package util

import (
	"fmt"

	"github.com/goccy/go-yaml/ast"
	"go.lsp.dev/protocol"
)

func GetNodeForPosition2(node ast.Node, position protocol.Position) *ast.Node {
	visitor := &PositionFinderVisitor{}
	ast.Walk(visitor, node)
	return nil
}

// PositionFinderVisitor is a visitor that collects positions.
type PositionFinderVisitor struct {
	result string // Field to store the result
	found  bool   // Flag to indicate if the target was found
}

// Visit method implementation for PositionFinderVisitor with a value receiver.
func (pv *PositionFinderVisitor) Visit(node ast.Node) ast.Visitor {
	// Example logic: if the node is a string and matches a condition, set found to true.
	strNode := node.String()
	fmt.Println(strNode)
	if strNode == "World" { // Example condition to stop the walk
		fmt.Println("Found the node")
		pv.found = true
		return nil // Stop the walk
	}
	pv.result += strNode + " "
	// Return a new instance of PositionFinderVisitor to continue traversal.
	return pv
}
