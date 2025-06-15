package util

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/stretchr/testify/assert"
)

// Unit test for PositionFinderVisitor
func TestPositionFinderVisitor(t *testing.T) {
	// Example AST represented as a slice of Nodes (strings).
	var node ast.Node

	yml := `
%YAML 1.2
---
a: 
  World: something
b: c
`

	err := yaml.Unmarshal([]byte(yml), &node)
	assert.NoError(t, err)

	// Create a PositionFinderVisitor instance.
	pv := &PositionFinderVisitor{}

	// Walk the AST with the position finder visitor.
	ast.Walk(pv, node)

	// Check if the result is as expected.
	expectedResult := "Hello "
	if pv.result != expectedResult {
		t.Errorf("Expected result: %q, got: %q", expectedResult, pv.result)
	}

	// Check if the visitor found the target.
	if !pv.found {
		t.Error("Expected to find the target node, but it was not found.")
	}
}
