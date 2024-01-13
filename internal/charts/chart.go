package charts

import (
	"github.com/mrjosh/helm-ls/internal/log"
	"go.lsp.dev/uri"
)

var logger = log.GetLogger()

type Chart struct {
	ValuesFiles   *ValuesFiles
	ChartMetadata *ChartMetadata
	RootURI       uri.URI
	ParentChart   ParentChart
}

func NewChart(rootURI uri.URI) *Chart {
	return &Chart{
		ValuesFiles:   NewValuesFiles(rootURI, "values.yaml", "values*.yaml"),
		ChartMetadata: NewChartMetadata(rootURI),
		RootURI:       rootURI,
		ParentChart:   getParentChart(rootURI),
	}
}

// ResolveValueFiles returns a list of all values files in the chart
// and all parent charts if the query tries to access global values
func (c *Chart) ResolveValueFiles(query []string, chartStore *ChartStore) []*ValuesFiles {
	if len(query) > 0 && query[0] == "global" {
		parentChart := c.ParentChart.GetParentChart(chartStore)
		if parentChart != nil {
			return append([]*ValuesFiles{c.ValuesFiles}, parentChart.ResolveValueFiles(query, chartStore)...)
		}
	}
	return []*ValuesFiles{c.ValuesFiles}
}
