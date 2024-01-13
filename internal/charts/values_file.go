package charts

import (
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/mrjosh/helm-ls/pkg/chartutil"
	"go.lsp.dev/uri"

	"gopkg.in/yaml.v3"
)

type ValuesFile struct {
	Values    chartutil.Values
	ValueNode yaml.Node
	URI       uri.URI
}

func NewValuesFile(filePath string) *ValuesFile {

	vals, err := chartutil.ReadValuesFile(filePath)
	if err != nil {
		logger.Error("Error loading values file ", filePath, err)
	}

	valueNodes, err := chartutil.ReadYamlFileToNode(filePath)
	if err != nil {
		logger.Error("Error loading values file ", filePath, err)
	}

	return &ValuesFile{
		ValueNode: valueNodes,
		Values:    vals,
		URI:       uri.New(util.FileURIScheme + filePath),
	}
}
