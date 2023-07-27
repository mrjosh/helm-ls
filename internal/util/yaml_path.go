package util

import (
	"fmt"
	"strings"
)

type YamlPath struct {
	TableNames []string
}

func NewYamlPath(yamlPathString string) (YamlPath, error) {
	var (
		splitted         = strings.Split(yamlPathString, ".")
		variableSplitted = []string{}
	)

	// filter out empty strings, that were added by the split
	for _, s := range splitted {
		if s != "" {
			variableSplitted = append(variableSplitted, s)
		}
	}
	if len(variableSplitted) == 0 {
		return YamlPath{}, fmt.Errorf("Could not parse yaml path: %s", yamlPathString)
	}
	// $ always points to the root context so we can safely remove it
	// as long the LSP does not know about ranges
	if variableSplitted[0] == "$" && len(variableSplitted) > 1 {
		variableSplitted = variableSplitted[1:]
	}

	return YamlPath{
		TableNames: variableSplitted,
	}, nil
}

func (path YamlPath) GetTail() []string {
	return path.TableNames[1:]
}

func (path YamlPath) IsValuesPath() bool {
	return path.TableNames[0] == "Values"
}

func (path YamlPath) IsChartPath() bool {
	return path.TableNames[0] == "Chart"
}

func (path YamlPath) IsReleasePath() bool {
	return path.TableNames[0] == "Release"
}

func (path YamlPath) IsFilesPath() bool {
	return path.TableNames[0] == "Files"
}
func (path YamlPath) IsCapabilitiesPath() bool {
	return path.TableNames[0] == "Capabilities"
}
