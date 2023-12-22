package lsp

import (
	"context"
	"testing"

	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
)

func TestGetVariableDefinitionDirectDecleration(t *testing.T) {

	var template = `
{{ $variable := "text" }}
{{ $variable }}
	`

	node, err := sitter.ParseCtx(context.Background(), []byte(template), gotemplate.GetLanguage())

	if err != nil {
		t.Errorf("Parsing did not work")
	}

	usageNode := node.NamedChild(1)
	definitionNode := GetVariableDefinitionOfNode(usageNode, template)

	if definitionNode == nil {
		t.Errorf("Could not get definitionNode")
	} else if definitionNode.Content([]byte(template)) != "$variable := \"text\"" {
		t.Errorf("Definition did not match but was %s", definitionNode.Content([]byte(template)))
	}
}

func TestGetVariableDefinitionOtherDecleration(t *testing.T) {
	var template = `
{{ $variable := "text" }}
{{ $someOther := "text" }}
{{ $variable }}
	`

	node, err := sitter.ParseCtx(context.Background(), []byte(template), gotemplate.GetLanguage())

	if err != nil {
		t.Errorf("Parsing did not work")
	}

	usageNode := node.NamedChild(2)
	definitionNode := GetVariableDefinitionOfNode(usageNode, template)

	if definitionNode == nil {
		t.Errorf("Could not get definitionNode")
	} else if definitionNode.Content([]byte(template)) != "$variable := \"text\"" {
		t.Errorf("Definition did not match but was %s", definitionNode.Content([]byte(template)))
	}

}

func TestGetVariableDefinitionRange(t *testing.T) {
	var template = `{{ range $index, $element := pipeline }}{{ $index }}{{ $element }}{{ end }}`
	// (template [0, 0] - [1, 0]
	//   (range_action [0, 0] - [0, 75]
	//     (range_variable_definition [0, 9] - [0, 37]
	//       index: (variable [0, 9] - [0, 15]
	//         name: (identifier [0, 10] - [0, 15]))
	//       element: (variable [0, 17] - [0, 25]
	//         name: (identifier [0, 18] - [0, 25]))
	//       range: (function_call [0, 29] - [0, 37]
	//         function: (identifier [0, 29] - [0, 37])))
	//     body: (variable [0, 43] - [0, 49]
	//       name: (identifier [0, 44] - [0, 49]))
	//     body: (variable [0, 55] - [0, 63]
	//       name: (identifier [0, 56] - [0, 63]))))

	node, err := sitter.ParseCtx(context.Background(), []byte(template), gotemplate.GetLanguage())

	if err != nil {
		t.Errorf("Parsing did not work")
	}

	elementUsageNode_start := sitter.Point{Row: 0, Column: 55}
	elementUsageNode := node.NamedDescendantForPointRange(elementUsageNode_start, elementUsageNode_start)
	if elementUsageNode == nil {
		t.Errorf("Could not get elementUsageNode")
	}
	if elementUsageNode.Content([]byte(template)) != "$element" {
		t.Errorf("Definition did not match but was %s", elementUsageNode.Content([]byte(template)))
	}
	definitionNode := GetVariableDefinitionOfNode(elementUsageNode, template)

	if definitionNode == nil {
		t.Errorf("Could not get definitionNode")
	} else if definitionNode.Content([]byte(template)) != "$element" {
		t.Errorf("Definition did not match but was %s", definitionNode.Content([]byte(template)))
	}

	indexUsageNode_start := sitter.Point{Row: 0, Column: 43}
	indexUsageNode := node.NamedDescendantForPointRange(indexUsageNode_start, indexUsageNode_start)
	if indexUsageNode == nil {
		t.Errorf("Could not get indexUsageNode")
	}
	if indexUsageNode.Content([]byte(template)) != "$index" {
		t.Errorf("Definition did not match but was %s", indexUsageNode.Content([]byte(template)))
	}
	definitionNode = GetVariableDefinitionOfNode(indexUsageNode, template)

	if definitionNode == nil {
		t.Errorf("Could not get definitionNode")
	} else if definitionNode.Content([]byte(template)) != "$index" {
		t.Errorf("Definition did not match but was %s", definitionNode.Content([]byte(template)))
	}
}
