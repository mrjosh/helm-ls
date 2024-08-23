package charts

import (
	"os"
	"path/filepath"

	"go.lsp.dev/uri"
)

func (c *Chart) GetDependecyURI(dependencyName string) uri.URI {
	unpackedPath := filepath.Join(c.RootURI.Filename(), "charts", dependencyName)
	fileInfo, err := os.Stat(unpackedPath)

	if err == nil && fileInfo.IsDir() {
		return uri.File(unpackedPath)
	}

	return uri.File(filepath.Join(c.RootURI.Filename(), "charts", DependencyCacheFolder, dependencyName))
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
