package charts

import (
	"fmt"
	"strings"

	"github.com/mrjosh/helm-ls/internal/log"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

var logger = log.GetLogger()

type Chart struct {
	ValuesFiles   *ValuesFiles
	ChartMetadata *ChartMetadata
	RootURI       uri.URI
	ParentChart   ParentChart
	HelmChart     *chart.Chart
}

func NewChart(rootURI uri.URI, valuesFilesConfig util.ValuesFilesConfig) *Chart {
	helmChart := loadHelmChart(rootURI)

	return &Chart{
		ValuesFiles: NewValuesFiles(rootURI,
			valuesFilesConfig.MainValuesFileName,
			valuesFilesConfig.LintOverlayValuesFileName,
			valuesFilesConfig.AdditionalValuesFilesGlobPattern),
		ChartMetadata: NewChartMetadata(rootURI),
		RootURI:       rootURI,
		ParentChart:   newParentChart(rootURI),
		HelmChart:     helmChart,
	}
}

func loadHelmChart(rootURI uri.URI) (helmChart *chart.Chart) {
	chartLoader, err := loader.Loader(rootURI.Filename())
	if err != nil {
		logger.Error(fmt.Sprintf("Error loading chart %s: %s", rootURI.Filename(), err.Error()))
		return nil
	}

	helmChart, err = chartLoader.Load()
	if err != nil {
		logger.Error(fmt.Sprintf("Error loading chart %s: %s", rootURI.Filename(), err.Error()))
	}

	if helmChart == nil {
		return &chart.Chart{}
	}

	return helmChart
}

func (c *Chart) GetMetadataLocation(templateContext []string) (lsp.Location, error) {
	modifyedVar := []string{}
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
