package charts

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/chart"
)

type DependencyTemplateFile struct {
	Content []byte
	Path    string
}

var DependencyCacheFolder = ".helm_ls_cache"

func (c *Chart) NewDependencyTemplateFile(chartName string, file *chart.File) *DependencyTemplateFile {
	path := filepath.Join(c.getDependencyDir(chartName), file.Name)

	return &DependencyTemplateFile{Content: file.Data, Path: path}
}

type PossibleDependencyFile interface {
	GetContent() []byte
	GetPath() string
}

func (c *Chart) getDependencyPathFromMetadata(chartName string) string {
	for _, dep := range c.ChartMetadata.Metadata.Dependencies {
		if dep.Name == chartName {
			return dep.Repository
		}
	}
	return ""
}

func (c *Chart) getDependencyDir(chartName string) string {
	dependencyURI := c.getDependencyPathFromMetadata(chartName)
	fileURIPrefix := "file://"
	if dependencyURI != "" && strings.HasPrefix(dependencyURI, fileURIPrefix) {
		relativePath := dependencyURI[len(fileURIPrefix):]
		relativePathClean, err := url.PathUnescape(relativePath)
		if err != nil {
			logger.Error("Could not unescape dependency file path", relativePath, err)
		}
		absolutePath := filepath.Join(c.RootURI.Filename(), relativePathClean)

		_, err = os.Stat(absolutePath)
		if err == nil {
			return absolutePath
		}
	}

	extractedPath := filepath.Join(c.RootURI.Filename(), "charts", chartName)
	_, err := os.Stat(extractedPath)
	if err == nil {
		return extractedPath
	}
	return filepath.Join(c.RootURI.Filename(), "charts", DependencyCacheFolder, chartName)
}

// SyncToDisk writes the content of the document to disk if it is a dependency file.
// If it is a dependency file, it was read from a archive, so we need to write it back,
// to be able to open it in a editor when using go-to-definition or go-to-reference.
func SyncToDisk(d PossibleDependencyFile) {
	if !IsDependencyFile(d) {
		return
	}
	err := os.MkdirAll(filepath.Dir(d.GetPath()), 0o755)
	if err == nil {
		err = os.WriteFile(d.GetPath(), []byte(d.GetContent()), 0o444)
	}
	if err != nil {
		logger.Error("Could not write dependency file", d.GetPath(), err)
	}
}

func IsDependencyFile(d PossibleDependencyFile) bool {
	return strings.Contains(d.GetPath(), DependencyCacheFolder)
}
