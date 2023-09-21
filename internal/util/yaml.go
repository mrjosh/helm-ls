package util

import (
	"fmt"

	lsp "go.lsp.dev/protocol"
	yamlv3 "gopkg.in/yaml.v3"
)

func GetPositionOfNode(node yamlv3.Node, query []string) (lsp.Position, error) {
	if node.IsZero() {
		return lsp.Position{}, fmt.Errorf("could not find Position of %s in values.yaml. Node was zero", query)
	}

	for index, value := range node.Content {
		if value.Value == "" {
			result, err := GetPositionOfNode(*value, query)
			if err == nil {
				return result, nil
			}
		}
		if value.Value == query[0] {
			if len(query) > 1 {
				if len(node.Content) < index+1 {
					return lsp.Position{}, fmt.Errorf("could not find Position of %s in values.yaml", query)
				}
				return GetPositionOfNode(*node.Content[index+1], query[1:])
			}
			return lsp.Position{Line: uint32(value.Line) - 1, Character: uint32(value.Column) - 1}, nil
		}
	}
	return lsp.Position{}, fmt.Errorf("could not find Position of %s in values.yaml. Found no match", query)

}
