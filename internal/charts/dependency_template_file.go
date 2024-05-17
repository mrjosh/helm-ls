package charts

import (
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart"
)

type DependencyTemplateFile struct {
	Content []byte
	Path    string
}

var DependencyCacheFolder = ".helm_ls_cache"

func (c *Chart) NewDependencyTemplateFile(chartName string, file *chart.File) *DependencyTemplateFile {
	path := filepath.Join(c.RootURI.Filename(), "charts", DependencyCacheFolder, chartName, file.Name)

	return &DependencyTemplateFile{Content: file.Data, Path: path}
}
