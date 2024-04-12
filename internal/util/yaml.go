package util

import (
	"fmt"
	"os"
	"strings"

	lsp "go.lsp.dev/protocol"
	yamlv3 "gopkg.in/yaml.v3"
)

func GetPositionOfNode(node *yamlv3.Node, query []string) (lsp.Position, error) {
	if node.IsZero() {
		return lsp.Position{}, fmt.Errorf("could not find Position of %s in values.yaml. Node was zero", query)
	}

	if len(query) == 0 {
		return lsp.Position{Line: uint32(node.Line) - 1, Character: uint32(node.Column) - 1}, nil
	}

	query[0] = strings.TrimSuffix(query[0], "[]")

	switch node.Kind {
	case yamlv3.DocumentNode:
		if len(node.Content) < 1 {
			return lsp.Position{}, fmt.Errorf("could not find Position of %s in values.yaml. Document is empty", query)
		}
		return GetPositionOfNode(node.Content[0], query)
	case yamlv3.SequenceNode:
		if len(node.Content) > 0 {
			return GetPositionOfNode(node.Content[0], query)
		}
	}

	for index, nestedNode := range node.Content {
		if nestedNode.Value == query[0] {
			if len(query) == 1 {
				return GetPositionOfNode(nestedNode, query[1:])
			}
			if len(node.Content) < index+1 {
				return lsp.Position{}, fmt.Errorf("could not find Position of %s in values.yaml", query)
			}
			return GetPositionOfNode(node.Content[index+1], query[1:])
		}
	}
	return lsp.Position{}, fmt.Errorf("could not find Position of %s in values.yaml. Found no match", query)
}

// ReadYamlFileToNode will parse a YAML file into a yaml Node.
func ReadYamlFileToNode(filename string) (node yamlv3.Node, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return yamlv3.Node{}, err
	}

	err = yamlv3.Unmarshal(data, &node)
	return node, err
}
