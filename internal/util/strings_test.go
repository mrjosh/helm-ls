package util

import (
	"testing"
)

func TestValueAtRemovesValueAfterDot(t *testing.T) {

	input := "test.go"
	expected := "test."

	result := ValueAt(input, 2)

	if expected != result {
		t.Errorf("Expected %s but got %s.", expected, result)
	}
}

func TestValueAtKeepsValueAfterDot(t *testing.T) {

	input := "test.go"
	expected := "test.go"

	result := ValueAt(input, 5)

	if expected != result {
		t.Errorf("Expected %s but got %s.", expected, result)
	}
}

func TestValueAtWithMoreContext(t *testing.T) {

	input := "test $.Values.test.go ____--------"
	expected := "$.Values.test."

	result := ValueAt(input, 15)

	if expected != result {
		t.Errorf("Expected %s but got %s.", expected, result)
	}
}
