package charts

import (
	"fmt"
	"os"

	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/uri"
	"helm.sh/helm/v3/pkg/chartutil"

	"gopkg.in/yaml.v3"
)

type ValuesFile struct {
	Values     chartutil.Values
	ValueNode  yaml.Node
	URI        uri.URI
	rawContent []byte
}

func NewValuesFileFromPath(filePath string) *ValuesFile {
	vals, valueNodes := readInValuesFile(filePath)

	return &ValuesFile{
		ValueNode: valueNodes,
		Values:    vals,
		URI:       uri.File(filePath),
	}
}

func NewValuesFileFromContent(uri uri.URI, data []byte) *ValuesFile {
	vals, valueNode := parseYaml(data)
	return &ValuesFile{
		ValueNode:  valueNode,
		Values:     vals,
		URI:        uri,
		rawContent: data,
	}
}

func (v *ValuesFile) Reload() {
	vals, valueNodes := readInValuesFile(v.URI.Filename())

	logger.Debug("Reloading values file", v.URI.Filename(), vals)
	v.Values = vals
	v.ValueNode = valueNodes
}

func readInValuesFile(filePath string) (chartutil.Values, yaml.Node) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error(fmt.Sprintf("Error loading values file %s ", filePath), err)
		return chartutil.Values{}, yaml.Node{}
	}

	return parseYaml(content)
}

func parseYaml(content []byte) (chartutil.Values, yaml.Node) {
	vals, err := chartutil.ReadValues(content)
	if err != nil {
		logger.Error("Error parsing values file ", err)
	}

	valueNodes, err := util.ReadYamlToNode(content)
	if err != nil {
		logger.Error("Error parsing values file ", err)
	}
	return vals, valueNodes
}

// GetContent implements PossibleDependencyFile.
func (v *ValuesFile) GetContent() []byte {
	return v.rawContent
}

// GetPath implements PossibleDependencyFile.
func (v *ValuesFile) GetPath() string {
	return v.URI.Filename()
}
