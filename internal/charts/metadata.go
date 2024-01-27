package charts

import (
	"path/filepath"

	"github.com/mrjosh/helm-ls/pkg/chart"
	"github.com/mrjosh/helm-ls/pkg/chartutil"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v3"
)

type ChartMetadata struct {
	YamlNode yaml.Node
	Metadata chart.Metadata
	URI      uri.URI
}

func NewChartMetadata(rootURI uri.URI) *ChartMetadata {
	filePath := filepath.Join(rootURI.Filename(), chartutil.ChartfileName)
	chartNode, err := chartutil.ReadYamlFileToNode(filePath)
	if err != nil {
		logger.Error("Error loading Chart.yaml file", rootURI, err)
	}

	return &ChartMetadata{
		Metadata: loadChartMetadata(filePath),
		YamlNode: chartNode,
		URI:      uri.File(filePath),
	}
}

func loadChartMetadata(filePath string) chart.Metadata {
	chartMetadata, err := chartutil.LoadChartfile(filePath)
	if err != nil {
		logger.Error("Error loading Chart.yaml file", filePath, err)
		return chart.Metadata{}
	}
	return *chartMetadata
}
