package charts

import (
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart"
)

type DependencyTemplateFile struct {
	Content []byte
	Path    string
}

func (d *DependencyTemplateFile) SyncToDisk() error {
	err := os.MkdirAll(filepath.Dir(d.Path), 0o755)
	if err == nil {
		err = os.WriteFile(d.Path, d.Content, 0o444)
	}
	if err != nil {
		logger.Error(err.Error())
	}

	return err
}

func (c *Chart) NewDeDependencyTemplateFile(chartName string, file *chart.File) *DependencyTemplateFile {
	path := filepath.Join(c.RootURI.Filename(), "charts", "helm_ls_cache", chartName, file.Name)

	return &DependencyTemplateFile{Content: file.Data, Path: path}
}
