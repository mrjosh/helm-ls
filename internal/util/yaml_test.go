package util

import (
	"fmt"
	lsp "go.lsp.dev/protocol"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGetPositionOfNode(t *testing.T) {

	data, err := os.ReadFile("./yaml_test_input.yaml")
	if err != nil {
		print(fmt.Sprint(err))
		t.Errorf("error yml parsing")
	}

	var node yaml.Node
	err = yaml.Unmarshal(data, &node)

	if err != nil {
		print(fmt.Sprint(err))
		t.Errorf("error yml parsing")
	}

	result, err := GetPositionOfNode(node, []string{"replicaCount"})
	expected := lsp.Position{Line: 6, Character: 1}
	if err != nil {
		t.Errorf("Result had error: %s", err)
	}
	if result != expected {
		t.Errorf("Result was not expected Position %v but was %v", expected, result)
	}

	result, err = GetPositionOfNode(node, []string{"image", "repository"})
	expected = lsp.Position{Line: 9, Character: 3}
	if err != nil {
		t.Errorf("Result had error: %s", err)
	}
	if result != expected {
		t.Errorf("Result was not expected Position %v but was %v", expected, result)
	}

	result, err = GetPositionOfNode(node, []string{"service", "test", "nested", "value"})
	expected = lsp.Position{Line: 31, Character: 7}
	if err != nil {
		t.Errorf("Result had error: %s", err)
	}
	if result != expected {
		t.Errorf("Result was not expected Position %v but was %v", expected, result)
	}
}
