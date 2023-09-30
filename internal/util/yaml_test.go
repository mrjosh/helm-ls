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
	expected := lsp.Position{Line: 5, Character: 0}
	if err != nil {
		t.Errorf("Result had error: %s", err)
	}
	if result != expected {
		t.Errorf("Result was not expected Position %v but was %v", expected, result)
	}

	result, err = GetPositionOfNode(node, []string{"image", "repository"})
	expected = lsp.Position{Line: 8, Character: 2}
	if err != nil {
		t.Errorf("Result had error: %s", err)
	}
	if result != expected {
		t.Errorf("Result was not expected Position %v but was %v", expected, result)
	}

	result, err = GetPositionOfNode(node, []string{"service", "test", "nested", "value"})
	expected = lsp.Position{Line: 30, Character: 6}
	if err != nil {
		t.Errorf("Result had error: %s", err)
	}
	if result != expected {
		t.Errorf("Result was not expected Position %v but was %v", expected, result)
	}
}
