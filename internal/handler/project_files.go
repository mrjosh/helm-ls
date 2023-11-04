package handler

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/mrjosh/helm-ls/pkg/chartutil"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

type ProjectFiles struct {
	ValuesFile string
	ChartFile  string
	RootURI    uri.URI
}

func (p ProjectFiles) GetValuesFileURI() lsp.DocumentURI {
	return "file://" + lsp.DocumentURI(p.ValuesFile)
}
func (p ProjectFiles) GetChartFileURI() lsp.DocumentURI {
	return "file://" + lsp.DocumentURI(p.ChartFile)
}

func NewProjectFiles(rootURI uri.URI, valuesFileName string) ProjectFiles {

	if valuesFileName == "" {
		valuesFileName = chartutil.ValuesfileName
	}
	valuesFileName = filepath.Join(rootURI.Filename(), valuesFileName)
	if _, err := os.Stat(valuesFileName); errors.Is(err, os.ErrNotExist) {
		valuesFileName = filepath.Join(rootURI.Filename(), "values.yml")
	}

	return ProjectFiles{
		ValuesFile: valuesFileName,
		ChartFile:  filepath.Join(rootURI.Filename(), chartutil.ChartfileName),
		RootURI:    rootURI,
	}
}
