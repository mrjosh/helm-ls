package charts

import (
	"os"
	"path/filepath"

	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
)

type ChartMetadata struct {
	YamlNode yaml.Node
	Metadata chart.Metadata
	URI      uri.URI
}

func NewChartMetadata(rootURI uri.URI) *ChartMetadata {
	filePath := filepath.Join(rootURI.Filename(), chartutil.ChartfileName)
	contents, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error("Error loading Chart.yaml file", filePath, err)
		return nil
	}

	chartNode, err := util.ReadYamlToNode(contents)
	if err != nil {
		logger.Error("Error loading Chart.yaml file", rootURI, err)
	}

	return &ChartMetadata{
		Metadata: loadChartMetadata(filePath),
		YamlNode: chartNode,
		URI:      uri.File(filePath),
	}
}

// Create a new ChartMetadata for a dependency chart, omitting the YamlNode since this is
// likely not required for dependency charts
func NewChartMetadataForDependencyChart(metadata *chart.Metadata, URI uri.URI) *ChartMetadata {
	return &ChartMetadata{
		Metadata: *metadata,
		URI:      URI,
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
