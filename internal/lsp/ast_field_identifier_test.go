package lsp

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
)

func TestGetFieldIdentifierPathSimple(t *testing.T) {
	template := `{{ .Values.test }}`

	var ast = ParseAst(template)
	// (template [0, 0] - [1, 0]
	//  (selector_expression [0, 3] - [0, 15]
	//    operand: (field [0, 3] - [0, 10]
	//      name: (identifier [0, 4] - [0, 10]))
	//    field: (field_identifier [0, 11] - [0, 15]))

	test_start := sitter.Point{Row: 0, Column: 12}
	testNode := ast.RootNode().NamedDescendantForPointRange(test_start, test_start)

	if testNode.Content([]byte(template)) != "test" {
		t.Errorf("Nodes were not correctly selected")
	}

	doc := Document{
		Content: template,
		Ast:     ast,
	}

	result := GetFieldIdentifierPath(testNode, &doc)
	assert.Equal(t, ".Values.test", result)
}

func TestGetFieldIdentifierPathWith(t *testing.T) {
	template := `{{ with .Values }}{{ .test }} {{ end }}`

	var ast = ParseAst(template)
	// (template [0, 0] - [1, 0]
	//  (with_action [0, 0] - [0, 39]
	//    condition: (field [0, 8] - [0, 15]
	//      name: (identifier [0, 9] - [0, 15]))
	//    consequence: (field [0, 21] - [0, 26]
	//      name: (identifier [0, 22] - [0, 26]))))

	test_start := sitter.Point{Row: 0, Column: 22}
	testNode := ast.RootNode().NamedDescendantForPointRange(test_start, test_start)

	if testNode.Content([]byte(template)) != "test" {
		t.Errorf("Nodes were not correctly selected")
	}

	doc := Document{
		Content: template,
		Ast:     ast,
	}

	result := GetFieldIdentifierPath(testNode, &doc)
	assert.Equal(t, ".Values.test", result)
}

func TestGetFieldIdentifierPathFunction(t *testing.T) {
	template := `{{ and .Values.test1 .Values.test2 }}`

	var ast = ParseAst(template)
	// (template [0, 0] - [1, 0]
	//   (function_call [0, 3] - [0, 35]
	//     function: (identifier [0, 3] - [0, 6])
	//     arguments: (argument_list [0, 7] - [0, 35]
	//       (selector_expression [0, 7] - [0, 20]
	//         operand: (field [0, 7] - [0, 14]
	//           name: (identifier [0, 8] - [0, 14]))
	//         field: (field_identifier [0, 15] - [0, 20]))
	//       (selector_expression [0, 21] - [0, 34]
	//         operand: (field [0, 21] - [0, 28]
	//           name: (identifier [0, 22] - [0, 28]))
	//         field: (field_identifier [0, 29] - [0, 34])))))
	//
	test1_start := sitter.Point{Row: 0, Column: 16}
	test2_start := sitter.Point{Row: 0, Column: 23}
	test1Node := ast.RootNode().NamedDescendantForPointRange(test1_start, test1_start)
	test2Node := ast.RootNode().NamedDescendantForPointRange(test2_start, test2_start)

	if test1Node.Content([]byte(template)) != "test1" || test2Node.Content([]byte(template)) != "test2" {
		t.Errorf("Nodes were not correctly selected")
	}

	doc := Document{
		Content: template,
		Ast:     ast,
	}

	// assert.Equal(t, ".Values.test1", GetFieldIdentifierPath(test1Node, &doc))
	assert.Equal(t, ".Values.test2", GetFieldIdentifierPath(test2Node, &doc))
}
