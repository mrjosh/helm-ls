package util

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"

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

	result, err := GetPositionOfNode(&node, []string{"replicaCount"})
	expected := lsp.Position{Line: 5, Character: 0}
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = GetPositionOfNode(&node, []string{"image", "repository"})
	expected = lsp.Position{Line: 8, Character: 2}
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = GetPositionOfNode(&node, []string{"service", "test", "nested", "value"})
	expected = lsp.Position{Line: 30, Character: 6}
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = GetPositionOfNode(&node, []string{"service", "test", "wrong", "value"})
	expected = lsp.Position{}
	assert.Error(t, err)
	assert.Equal(t, expected, result)
}

func TestGetPositionOfNodeWithList(t *testing.T) {
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

	result, err := GetPositionOfNode(&node, []string{"list[0]"})
	expected := lsp.Position{Line: 32, Character: 0}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = GetPositionOfNode(&node, []string{"list[0]", "first"})
	expected = lsp.Position{Line: 33, Character: 4}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	result, err = GetPositionOfNode(&node, []string{"notExistingList[0]", "first"})
	expected = lsp.Position{}

	assert.Error(t, err)
	assert.Equal(t, expected, result)
}

func TestGetPositionOfNodeInEmptyDocument(t *testing.T) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte(""), &node)
	if err != nil {
		print(fmt.Sprint(err))
		t.Errorf("error yml parsing")
	}

	result, err := GetPositionOfNode(&node, []string{"list[0]"})
	expected := lsp.Position{}

	assert.Error(t, err)
	assert.Equal(t, expected, result)

	err = yaml.Unmarshal([]byte("  "), &node)
	if err != nil {
		print(fmt.Sprint(err))
		t.Errorf("error yml parsing")
	}

	result, err = GetPositionOfNode(&node, []string{"list[0]"})
	expected = lsp.Position{}

	assert.Error(t, err)
	assert.Equal(t, expected, result)
}
