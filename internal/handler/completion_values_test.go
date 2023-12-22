package handler

import (
	"testing"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"go.lsp.dev/protocol"
)

func TestEmptyValues(t *testing.T) {
	handler := &langHandler{
		linterName: "helm-lint",
		connPool:   nil,
		documents:  nil,
		values:     make(map[string]interface{}),
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
		values:     make(map[string]interface{}),
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
		values:     make(map[string]interface{}),
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
