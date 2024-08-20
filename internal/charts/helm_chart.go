package charts

import (
	"path/filepath"

	"go.lsp.dev/uri"
	"helm.sh/helm/v3/pkg/chart"
)

// TODO: this ignores the newChart callback present in the ChartStore
func NewChartFromHelmChart(helmChart *chart.Chart, rootURI uri.URI) *Chart {
	valuesFile := getValues(helmChart, rootURI)
	return &Chart{
		ValuesFiles: &ValuesFiles{
			MainValuesFile:        valuesFile,
			OverlayValuesFile:     &ValuesFile{},
			AdditionalValuesFiles: []*ValuesFile{},
		},
		ChartMetadata: NewChartMetadataForDependencyChart(helmChart.Metadata, rootURI),
		RootURI:       rootURI,
		ParentChart:   ParentChart{},
		HelmChart:     helmChart,
	}
}

func getValues(helmChart *chart.Chart, rootURI uri.URI) *ValuesFile {
	// Use Raw values if present because they also contain comments and documentation can be useful
	for _, file := range helmChart.Raw {
		if file.Name == "values.yaml" {
			return NewValuesFileFromContent(uri.File(filepath.Join(rootURI.Filename(), "values.yaml")), file.Data)
		}
	}
	return NewValuesFileFromValues(uri.File(filepath.Join(rootURI.Filename(), "values.yaml")), helmChart.Values)
}
