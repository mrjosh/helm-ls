package charts

import (
	"fmt"

	"go.lsp.dev/uri"
)

type QueriedValuesFiles struct {
	Selector    []string
	ValuesFiles *ValuesFiles
}

// ResolveValueFiles returns a list of all values files in the chart
// and all parent and dependency charts with the adjusted query
// that match a given query
// a query starting with "global" returns all charts
// a query starting with "subchartName" will return the subchart if it is named "subchartName"
func (c *Chart) ResolveValueFiles(query []string, chartStore *ChartStore) (result []*QueriedValuesFiles) {
	recResult := map[uri.URI]*QueriedValuesFiles{}
	c.resolveValueFilesRecursive(query, chartStore, recResult)

	// TODO: @qvalentin use maps.Values once we have Go 1.23
	for _, valuesFiles := range recResult {
		result = append(result, valuesFiles)
	}
	return result
}

func (c *Chart) resolveValueFilesRecursive(query []string, chartStore *ChartStore, result map[uri.URI]*QueriedValuesFiles) {
	// check if chart was already processed
	if _, ok := result[c.RootURI]; ok {
		return
	}

	if c == nil {
		logger.Error("Could not resolve values files for nil chart")
		return
	}

	currentResult := &QueriedValuesFiles{Selector: query, ValuesFiles: c.ValuesFiles}
	result[c.RootURI] = currentResult
	if len(query) == 0 {
		return
	}

	c.resolveValuesFilesOfDependencies(query, chartStore, result)
	c.resolveValuesFilesOfParent(chartStore, query, result)
}

func (c *Chart) resolveValuesFilesOfParent(chartStore *ChartStore, query []string, result map[uri.URI]*QueriedValuesFiles) {
	parentChart := c.ParentChart.GetParentChart(chartStore)
	if parentChart == nil {
		return
	}

	if query[0] == "global" {
		parentChart.resolveValueFilesRecursive(query, chartStore, result)
	}

	chartName := c.ChartMetadata.Metadata.Name
	extendedQuery := append([]string{chartName}, query...)

	parentChart.resolveValueFilesRecursive(extendedQuery, chartStore, result)
}

func (c *Chart) resolveValuesFilesOfDependencies(query []string, chartStore *ChartStore, result map[uri.URI]*QueriedValuesFiles) {
	if c.HelmChart == nil {
		return
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

			dependencyChart.resolveValueFilesRecursive(subQuery, chartStore, result)
		}
	}
}
