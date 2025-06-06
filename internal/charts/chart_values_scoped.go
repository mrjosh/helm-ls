package charts

import (
	"fmt"
)

type ScopedValuesFiles struct {
	// Scope defines a scope of the values files within the current chart.
	// Scope is used for parent charts, it indicates the scope in the parents charts values where the child chart values are relevant
	Scope []string
	// SubScope defines a scope within the values files. "global" is always an implicit subscope.
	// SubScope is used for dependency charts to indicate the scope of the parents values that is relevant for the current chart
	SubScope    []string
	ValuesFiles *ValuesFiles
	Chart       *Chart
}

// For a given chart, return all values files
// return values files of parents if they contain values matching the current chartname
// or global values
//
// return values files of dependencies with the shifted scope
// e.g. a subchart called subchart should be returned with the scope subchart (nested subcharts should have the scope subchart/subchart)
func (c *Chart) GetScopedValuesFiles(chartStore *ChartStore) []*ScopedValuesFiles {
	result := []*ScopedValuesFiles{
		{Scope: []string{}, SubScope: []string{}, ValuesFiles: c.ValuesFiles, Chart: c},
	}

	result = append(result, c.GetScopedValuesFileDependencies(chartStore)...)
	return append(result, c.GetScopedValuesFileParents(chartStore)...)
}

func (c *Chart) GetScopedValuesFileParents(chartStore *ChartStore) []*ScopedValuesFiles {
	result := []*ScopedValuesFiles{}

	if !c.ParentChart.HasParent {
		return result
	}

	parent := c.ParentChart.GetParentChart(chartStore)
	if parent == nil {
		return result
	}

	ownParentResult := parent.ValuesFiles

	result = append(result, &ScopedValuesFiles{
		Scope: []string{}, SubScope: []string{
			c.Name(),
		},
		ValuesFiles: ownParentResult,
		Chart:       parent,
	})

	recResult := parent.GetScopedValuesFileParents(chartStore)

	for i := range recResult {
		recResult[i].SubScope = append(recResult[i].SubScope, c.Name())
	}

	result = append(result, recResult...)
	return result
}

func (c *Chart) GetScopedValuesFileDependencies(chartStore *ChartStore) []*ScopedValuesFiles {
	result := []*ScopedValuesFiles{}

	for _, dependency := range c.HelmChart.Dependencies() {
		dependencyChart := chartStore.Charts[c.GetDependecyURI(dependency.Name())]
		if dependencyChart == nil {
			logger.Error(fmt.Sprintf("Could not find dependency %s", dependency.Name()))
			continue
		}

		dependencyResult := dependencyChart.ValuesFiles
		result = append(result, &ScopedValuesFiles{
			Scope:       []string{dependency.Name()},
			SubScope:    []string{},
			ValuesFiles: dependencyResult,
			Chart:       dependencyChart,
		})

		recResult := dependencyChart.GetScopedValuesFileDependencies(chartStore)

		for i := range recResult {
			recResult[i].Scope = append([]string{dependency.Name()}, recResult[i].Scope...)
		}
		result = append(result, recResult...)
	}

	return result
}
