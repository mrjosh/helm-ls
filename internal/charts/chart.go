package charts

import (
	"strings"

	"github.com/mrjosh/helm-ls/internal/log"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
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

func (c *Chart) GetValueLocation(templateContext []string) (lsp.Location, error) {
	modifyedVar := make([]string, len(templateContext))
	// make the first letter lowercase since in the template the first letter is
	// capitalized, but it is not in the Chart.yaml file
	for _, value := range templateContext {
		restOfString := ""
		if (len(value)) > 1 {
			restOfString = value[1:]
		}
		firstLetterLowercase := strings.ToLower(string(value[0])) + restOfString
		modifyedVar = append(modifyedVar, firstLetterLowercase)
	}
	position, err := util.GetPositionOfNode(&c.ChartMetadata.YamlNode, modifyedVar)

	return lsp.Location{URI: c.ChartMetadata.URI, Range: lsp.Range{Start: position}}, err
}
