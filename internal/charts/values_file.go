package charts

import (
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
