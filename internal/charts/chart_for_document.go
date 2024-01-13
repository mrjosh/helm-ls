package charts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrjosh/helm-ls/pkg/chartutil"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (s *ChartStore) GetChartForDoc(uri lsp.DocumentURI) (*Chart, error) {
	chart := s.getChartFromCache(uri)
	if chart != nil {
		return chart, nil
	}

	chart = s.getChartFromFilesystemForTemplates(uri.Filename())

	if chart != nil {
		s.Charts[chart.RootURI] = chart
		return chart, nil
	}

	return nil, ErrChartNotFound{
		URI: uri,
	}
}

func (s *ChartStore) getChartFromCache(uri lsp.DocumentURI) *Chart {
	for chartURI, chart := range s.Charts {
		if strings.HasPrefix(uri.Filename(), filepath.Join(chartURI.Filename(), "template")) {
			return chart
		}
	}
	return nil
}

func (s *ChartStore) getChartFromFilesystemForTemplates(path string) *Chart {
	directory := filepath.Dir(path)
	if filepath.Base(directory) == "templates" {
		templatesDir := directory
		expectedChartDir := filepath.Dir(templatesDir)

		// check if Chart.yaml exists
		if isChartDirectory(expectedChartDir) {
			return s.newChart(uri.New("file://" + expectedChartDir))
		}
	}

	rootDirectory := s.RootURI.Filename()
	if directory == rootDirectory {
		return nil
	}

	return s.getChartFromFilesystemForTemplates(directory)
}

func isChartDirectory(expectedChartDir string) bool {
	_, err := os.Stat(filepath.Join(expectedChartDir, chartutil.ChartfileName))
	return err == nil
}

type ErrChartNotFound struct {
	URI lsp.DocumentURI
}

func (e ErrChartNotFound) Error() string {
	return fmt.Sprintf("Chart not found for file: %s", e.URI)
}
