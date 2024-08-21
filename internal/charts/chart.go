package charts

import (
	"fmt"
	"os"
	"path/filepath"
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

func (c *Chart) GetDependecyURI(dependencyName string) uri.URI {
	unpackedPath := filepath.Join(c.RootURI.Filename(), "charts", dependencyName)
	fileInfo, err := os.Stat(unpackedPath)

	if err == nil && fileInfo.IsDir() {
		return uri.File(unpackedPath)
	}

	return uri.File(filepath.Join(c.RootURI.Filename(), "charts", DependencyCacheFolder, dependencyName))
}

func loadHelmChart(rootURI uri.URI) *chart.Chart {
	var helmChart *chart.Chart

	loader, err := loader.Loader(rootURI.Filename())
	if err != nil {
		logger.Error(fmt.Sprintf("Error loading chart %s: %s", rootURI.Filename(), err.Error()))
		return nil
	}

	helmChart, err = loader.Load()
	if err != nil {
		logger.Error(fmt.Sprintf("Error loading chart %s: %s", rootURI.Filename(), err.Error()))
	}

	return helmChart
}

type QueriedValuesFiles struct {
	Selector    []string
	ValuesFiles *ValuesFiles
}

// ResolveValueFiles returns a list of all values files in the chart
// and all parent charts if the query tries to access global values
func (c *Chart) ResolveValueFiles(query []string, chartStore *ChartStore) []*QueriedValuesFiles {
	if c == nil {
		logger.Error("Could not resolve values files for nil chart")
		return []*QueriedValuesFiles{}
	}
	result := []*QueriedValuesFiles{{Selector: query, ValuesFiles: c.ValuesFiles}}
	if len(query) == 0 {
		return result
	}

	result = append(result, c.resolveValuesFilesOfDependencies(query, chartStore)...)

	result = append(result, c.resolveValuesFilesOfParent(chartStore, query)...)

	return result
}

func (c *Chart) resolveValuesFilesOfParent(chartStore *ChartStore, query []string) (result []*QueriedValuesFiles) {
	parentChart := c.ParentChart.GetParentChart(chartStore)
	if parentChart == nil {
		return []*QueriedValuesFiles{}
	}

	if query[0] == "global" {
		return append(result,
			parentChart.ResolveValueFiles(query, chartStore)...)
	}

	chartName := c.ChartMetadata.Metadata.Name
	extendedQuery := append([]string{chartName}, query...)

	return append(result,
		parentChart.ResolveValueFiles(extendedQuery, chartStore)...)
}

func (c *Chart) resolveValuesFilesOfDependencies(query []string, chartStore *ChartStore) (result []*QueriedValuesFiles) {
	if c.HelmChart == nil {
		return []*QueriedValuesFiles{}
	}
	for _, dependency := range c.HelmChart.Dependencies() {
		logger.Debug(fmt.Sprintf("Resolving dependency %s with query %s", dependency.Name(), query))

		if dependency.Name() == query[0] || query[0] == "global" {
			subQuery := query

			if dependency.Name() == query[0] {
				if len(query) >= 1 {
					subQuery = query[1:]
				}
			}

			dependencyChart := chartStore.Charts[c.GetDependecyURI(dependency.Name())]
			if dependencyChart == nil {
				logger.Error(fmt.Sprintf("Could not find dependency %s", dependency.Name()))
				continue
			}

			// TODO: why does this cause infinite recursion?
			// because dependencyChart have the parent as dependency
			// result = append(result, dependencyChart.ResolveValueFiles(subQuery, chartStore)...)

			result = append(result,
				&QueriedValuesFiles{Selector: subQuery, ValuesFiles: dependencyChart.ValuesFiles})
		}
	}

	return result
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

func (c *Chart) GetDependeciesTemplates() []*DependencyTemplateFile {
	result := []*DependencyTemplateFile{}
	if c.HelmChart == nil {
		return result
	}

	for _, dependency := range c.HelmChart.Dependencies() {
		for _, file := range dependency.Templates {
			dependencyTemplate := c.NewDependencyTemplateFile(dependency.Name(), file)
			result = append(result, dependencyTemplate)
		}
	}

	return result
}
