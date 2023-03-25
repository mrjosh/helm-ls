package handler

import (
	"testing"
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
