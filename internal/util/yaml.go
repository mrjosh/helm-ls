package util

import (
	"fmt"
	"strings"

	lsp "go.lsp.dev/protocol"
	yamlv3 "gopkg.in/yaml.v3"
)

func GetPositionOfNode(node *yamlv3.Node, query []string) (lsp.Position, error) {
	if node.IsZero() {
		return lsp.Position{}, fmt.Errorf("could not find Position of %s in values.yaml. Node was zero", query)
	}

	if node.Kind == yamlv3.DocumentNode {
		if len(node.Content) < 1 {
			return lsp.Position{}, fmt.Errorf("could not find Position of %s in values.yaml. Document is empty", query)
		}
		return GetPositionOfNode(node.Content[0], query)
	}

	if len(query) == 0 {
		return lsp.Position{Line: uint32(node.Line) - 1, Character: uint32(node.Column) - 1}, nil
	}

	isRange := false

	if strings.HasSuffix(query[0], "[]") {
		query = append([]string{}, query...)
		query[0] = strings.TrimSuffix(query[0], "[]")
		isRange = true
	}

	kind := node.Kind
	switch kind {
	case yamlv3.SequenceNode:
		if len(node.Content) > 0 {
			return GetPositionOfNode(node.Content[0], query)
		}
	}

	checkNested := []string{}
	for index, nestedNode := range node.Content {
		checkNested = append(checkNested, nestedNode.Value)
		if nestedNode.Value == query[0] {
			if len(query) == 1 {
				return GetPositionOfNode(nestedNode, query[1:])
			}
			if len(node.Content) < index+1 {
				return lsp.Position{}, fmt.Errorf("could not find Position of %s in values", query)
			}
			if isRange {
				return getPositionOfNodeAfterRange(node.Content[index+1], query[1:])
			}
			return GetPositionOfNode(node.Content[index+1], query[1:])
		}
	}
	return lsp.Position{}, fmt.Errorf("could not find Position of %s in values.yaml. Found no match. Possible values %v. Kind is %d", query, checkNested, kind)
}

func getPositionOfNodeAfterRange(node *yamlv3.Node, query []string) (lsp.Position, error) {
	switch node.Kind {
	case yamlv3.SequenceNode:
		if len(node.Content) > 0 {
			return GetPositionOfNode(node.Content[0], query)
		}
	case yamlv3.MappingNode:
		if len(node.Content) > 1 {
			return GetPositionOfNode(node.Content[1], query)
		}
	}

	return lsp.Position{}, fmt.Errorf("could not find Position of %s in values. Found no match", query)
}

func GetNodeForPosition(node *yamlv3.Node, position lsp.Position) *yamlv3.Node {
	if node.IsZero() {
		return nil
	}

	if node.Value != "" && node.Line == int(position.Line+1) && node.Column <= int(position.Character+1) {
		return node
	}

	for _, nestedNode := range node.Content {
		nestedResult := GetNodeForPosition(nestedNode, position)
		if nestedResult != nil {
			return nestedResult
		}
	}

	return nil
}

// ReadYamlToNode will parse a YAML file into a yaml Node.
func ReadYamlToNode(data []byte) (node yamlv3.Node, err error) {
	err = yamlv3.Unmarshal(data, &node)
	return node, err
}
