package charts

import (
	"fmt"

	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/uri"
	"helm.sh/helm/v3/pkg/chartutil"

	"gopkg.in/yaml.v3"
)

type ValuesFile struct {
	Values    chartutil.Values
	ValueNode yaml.Node
	URI       uri.URI
}

func NewValuesFile(filePath string) *ValuesFile {
	vals, valueNodes := readInValuesFile(filePath)

	return &ValuesFile{
		ValueNode: valueNodes,
		Values:    vals,
		URI:       uri.File(filePath),
	}
}

func NewValuesFileFromValues(uri uri.URI, values chartutil.Values) *ValuesFile {
	valueNode, error := util.ValuesToYamlNode(values)
	if error != nil {
		logger.Error(fmt.Sprintf("Could not load values for file %s", uri.Filename()), error)
		return &ValuesFile{}
	}
	return &ValuesFile{
		ValueNode: valueNode,
		Values:    values,
		URI:       uri,
	}
}

func (v *ValuesFile) Reload() {
	vals, valueNodes := readInValuesFile(v.URI.Filename())

	logger.Debug("Reloading values file", v.URI.Filename(), vals)
	v.Values = vals
	v.ValueNode = valueNodes
}

func readInValuesFile(filePath string) (chartutil.Values, yaml.Node) {
	vals, err := chartutil.ReadValuesFile(filePath)
	if err != nil {
		logger.Error("Error loading values file ", filePath, err)
	}

	valueNodes, err := util.ReadYamlFileToNode(filePath)
	if err != nil {
		logger.Error("Error loading values file ", filePath, err)
	}
	return vals, valueNodes
}

// GetContent implements PossibleDependencyFile.
func (d *ValuesFile) GetContent() string {
	yaml, err := yaml.Marshal(d.Values)
	if err != nil {
		logger.Error(fmt.Sprintf("Could not load values for file %s", d.URI.Filename()), err)
		return ""
	}
	return string(yaml)
}

// GetPath implements PossibleDependencyFile.
func (d *ValuesFile) GetPath() string {
	return d.URI.Filename()
}
