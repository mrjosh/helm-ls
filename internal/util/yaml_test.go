package util

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"

	"gopkg.in/yaml.v3"
)

func TestGetPositionOfNodeInEmptyDocument(t *testing.T) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte(""), &node)
	if err != nil {
		print(fmt.Sprint(err))
		t.Errorf("error yml parsing")
	}

	result, err := GetPositionOfNode(&node, []string{"list[]"})
	expected := lsp.Position{}

	assert.Error(t, err)
	assert.Equal(t, expected, result)

	err = yaml.Unmarshal([]byte("  "), &node)
	if err != nil {
		print(fmt.Sprint(err))
		t.Errorf("error yml parsing")
	}

	result, err = GetPositionOfNode(&node, []string{"list[]"})
	expected = lsp.Position{}

	assert.Error(t, err)
	assert.Equal(t, expected, result)
}

func TestGetPositionOfNodeTable(t *testing.T) {
	tests := []struct {
		name      string
		query     []string
		expected  lsp.Position
		expectErr bool
	}{
		{"replicaCount", []string{"replicaCount"}, lsp.Position{Line: 5, Character: 0}, false},
		{"image, repository", []string{"image", "repository"}, lsp.Position{Line: 8, Character: 2}, false},
		{"service, test, nested, value", []string{"service", "test", "nested", "value"}, lsp.Position{Line: 30, Character: 6}, false},
		{"service, test, wrong, value", []string{"service", "test", "wrong", "value"}, lsp.Position{}, true},
		{"list[]", []string{"list[]"}, lsp.Position{Line: 32, Character: 0}, false},
		{"list[], first", []string{"list[]", "first"}, lsp.Position{Line: 33, Character: 4}, false},
		{"notExistingList[], first", []string{"notExistingList[]", "first"}, lsp.Position{}, true},
		{"mapping[], something", []string{"mapping[]", "something"}, lsp.Position{Line: 40, Character: 4}, false},
	}

	data, err := os.ReadFile("./yaml_test_input.yaml")
	if err != nil {
		t.Fatalf("error reading test input file: %v", err)
	}

	var node yaml.Node
	err = yaml.Unmarshal(data, &node)
	if err != nil {
		t.Fatalf("error parsing YAML: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryCopy := append([]string{}, tt.query...)

			result, err := GetPositionOfNode(&node, tt.query)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			assert.Equal(t, queryCopy, tt.query)
		})
	}
}
