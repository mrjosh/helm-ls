package charts

import (
	"go.lsp.dev/uri"
)

func (c *Chart) GetDependecyURI(dependencyName string) uri.URI {
	return uri.File(c.getDependencyDir(dependencyName))
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
