package charts

import (
	"github.com/mrjosh/helm-ls/internal/log"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/uri"
)

var logger = log.GetLogger()

type Chart struct {
	ValuesFiles   *ValuesFiles
	ChartMetadata *ChartMetadata
	RootURI       uri.URI
	ParentChart   ParentChart
}

func NewChart(rootURI uri.URI, valuesFilesConfig util.ValuesFilesConfig) *Chart {
	return &Chart{
		ValuesFiles: NewValuesFiles(rootURI,
			valuesFilesConfig.MainValuesFileName,
			valuesFilesConfig.LintOverlayValuesFileName,
			valuesFilesConfig.AdditionalValuesFilesGlobPattern),
		ChartMetadata: NewChartMetadata(rootURI),
		RootURI:       rootURI,
		ParentChart:   newParentChart(rootURI),
	}
}

type QueriedValuesFiles struct {
	Selector    []string
	ValuesFiles *ValuesFiles
}

// ResolveValueFiles returns a list of all values files in the chart
// and all parent charts if the query tries to access global values
func (c *Chart) ResolveValueFiles(query []string, chartStore *ChartStore) []*QueriedValuesFiles {
	ownResult := []*QueriedValuesFiles{{Selector: query, ValuesFiles: c.ValuesFiles}}
	if len(query) == 0 {
		return ownResult
	}

	parentChart := c.ParentChart.GetParentChart(chartStore)
	if parentChart == nil {
		return ownResult
	}

	if query[0] == "global" {
		return append(ownResult,
			parentChart.ResolveValueFiles(query, chartStore)...)
	}

	chartName := c.ChartMetadata.Metadata.Name
	extendedQuery := append([]string{chartName}, query...)
	return append(ownResult,
		parentChart.ResolveValueFiles(extendedQuery, chartStore)...)
}
