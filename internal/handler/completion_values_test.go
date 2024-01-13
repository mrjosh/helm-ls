package handler

import (
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
	"gopkg.in/yaml.v3"
)

func TestEmptyValues(t *testing.T) {
	handler := &langHandler{
		linterName: "helm-lint",
		connPool:   nil,
		documents:  nil,
	}

	var result = handler.getValue(make(map[string]interface{}), []string{"global"})

	if len(result) != 0 {
		t.Errorf("Length of result was not zero.")
	}
	result = handler.getValue(make(map[string]interface{}), []string{""})

	if len(result) != 0 {
		t.Errorf("Length of result was not zero.")
	}
}

func TestValues(t *testing.T) {
	handler := &langHandler{
		linterName: "helm-lint",
		connPool:   nil,
		documents:  nil,
	}
	var nested = map[string]interface{}{"nested": "value"}
	var values = map[string]interface{}{"global": nested}

	result := handler.getValue(values, []string{"g"})

	if len(result) != 1 || result[0].InsertText != "global" {
		t.Errorf("Completion for g was not global but was %s.", result[0].InsertText)
	}

	result = handler.getValue(values, []string{""})

	if len(result) != 1 || result[0].InsertText != "global" {
		t.Errorf("Completion for \"\" was not global but was %s.", result[0].InsertText)
	}

	result = handler.getValue(values, []string{"global", "nes"})
	if len(result) != 1 || result[0].InsertText != "nested" {
		t.Errorf("Completion for global.nes was not nested but was %s.", result[0].InsertText)
	}
}

func TestWrongValues(t *testing.T) {
	handler := &langHandler{
		linterName: "helm-lint",
		connPool:   nil,
		documents:  nil,
	}
	var nested = map[string]interface{}{"nested": 1}
	var values = map[string]interface{}{"global": nested}

	result := handler.getValue(values, []string{"some", "wrong", "values"})
	if len(result) != 0 {
		t.Errorf("Length of result was not zero.")
	}

	result = handler.getValue(values, []string{"some", "wrong"})
	if len(result) != 0 {
		t.Errorf("Length of result was not zero.")
	}

	result = handler.getValue(values, []string{"some", ""})
	if len(result) != 0 {
		t.Errorf("Length of result was not zero.")
	}

	result = handler.getValue(values, []string{"global", "nested", ""})
	if len(result) != 0 {
		t.Errorf("Length of result was not zero.")
	}
}

func TestCompletionAstParsing(t *testing.T) {

	documentText := `{{ .Values.global. }}`
	expectedWord := ".Values.global."
	doc := &lsplocal.Document{
		Content: documentText,
		Ast:     lsplocal.ParseAst(documentText),
	}
	position := protocol.Position{
		Line:      0,
		Character: 18,
	}
	word, _ := completionAstParsing(doc, position)
	if expectedWord != word {
		t.Errorf("Expected word '%s', but got '%s'", expectedWord, word)
	}

}

func TestGetValuesCompletions(t *testing.T) {
	handler := &langHandler{
		linterName: "helm-lint",
		connPool:   nil,
		documents:  nil,
	}
	var nested = map[string]interface{}{"nested": "value"}
	var valuesMain = map[string]interface{}{"global": nested}
	var valuesAdditional = map[string]interface{}{"glob": nested}
	chart := &charts.Chart{
		ValuesFiles: &charts.ValuesFiles{
			MainValuesFile: &charts.ValuesFile{
				Values:    valuesMain,
				ValueNode: yaml.Node{},
				URI:       "",
			},
			AdditionalValuesFiles: []*charts.ValuesFile{
				{
					Values:    valuesAdditional,
					ValueNode: yaml.Node{},
					URI:       "",
				},
			},
		},
		RootURI: "",
	}

	result := handler.getValuesCompletions(chart, []string{"g"})
	assert.Equal(t, 2, len(result))

	result = handler.getValuesCompletions(chart, []string{"something", "different"})
	assert.Empty(t, result)
}

func TestGetValuesCompletionsContainsNoDupliactes(t *testing.T) {
	handler := &langHandler{
		linterName: "helm-lint",
		connPool:   nil,
		documents:  nil,
	}
	var nested = map[string]interface{}{"nested": "value"}
	var valuesMain = map[string]interface{}{"global": nested}
	var valuesAdditional = map[string]interface{}{"global": nested}
	chart := &charts.Chart{
		ValuesFiles: &charts.ValuesFiles{
			MainValuesFile: &charts.ValuesFile{
				Values:    valuesMain,
				ValueNode: yaml.Node{},
				URI:       "",
			},
			AdditionalValuesFiles: []*charts.ValuesFile{
				{
					Values: valuesAdditional,
					URI:    "",
				},
			},
		},
		RootURI: "",
	}

	result := handler.getValuesCompletions(chart, []string{"g"})
	assert.Equal(t, 1, len(result))
}
