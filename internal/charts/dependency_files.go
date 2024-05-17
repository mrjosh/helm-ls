package charts

import (
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
	path := filepath.Join(c.RootURI.Filename(), "charts", DependencyCacheFolder, chartName, file.Name)

	return &DependencyTemplateFile{Content: file.Data, Path: path}
}

type PossibleDependencyFile interface {
	GetContent() string
	GetPath() string
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
