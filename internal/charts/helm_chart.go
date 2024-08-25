package charts

import (
	"path/filepath"

	"go.lsp.dev/uri"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
)

func NewChartFromHelmChart(helmChart *chart.Chart, rootURI uri.URI) *Chart {
	return &Chart{
		ValuesFiles: &ValuesFiles{
			MainValuesFile:        getValues(helmChart, rootURI),
			OverlayValuesFile:     &ValuesFile{},
			AdditionalValuesFiles: []*ValuesFile{},
		},
		ChartMetadata: NewChartMetadataForDependencyChart(helmChart.Metadata, rootURI),
		RootURI:       rootURI,
		ParentChart:   getParent(helmChart, rootURI),
		HelmChart:     helmChart,
	}
}

func getValues(helmChart *chart.Chart, rootURI uri.URI) *ValuesFile {
	// Use Raw values if present because they also contain comments and documentation can be useful
	uri := uri.File(filepath.Join(rootURI.Filename(), "values.yaml"))
	for _, file := range helmChart.Raw {
		if file.Name == "values.yaml" {
			return NewValuesFileFromContent(uri, file.Data)
		}
	}
	return &ValuesFile{
		ValueNode:  yaml.Node{},
		Values:     helmChart.Values,
		URI:        uri,
		rawContent: []byte{},
	}
}

func getParent(helmChart *chart.Chart, rootURI uri.URI) ParentChart {
	if helmChart.Parent() != nil {
		return newParentChart(rootURI)
	}
	return ParentChart{}
}
