package charts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"helm.sh/helm/v3/pkg/chartutil"
)

func (s *ChartStore) GetChartForDoc(uri lsp.DocumentURI) (*Chart, error) {
	chart := s.getChartFromCache(uri)
	if chart != nil {
		return chart, nil
	}

	chart, err := s.getChartFromFilesystemForTemplates(uri.Filename())
	if err != nil {
		return chart, ErrChartNotFound{
			URI: uri,
		}
	}
	s.AddChart(chart)

	return chart, nil
}

func (s *ChartStore) GetChartOrParentForDoc(uri lsp.DocumentURI) (*Chart, error) {
	chart, err := s.GetChartForDoc(uri)
	if err != nil {
		return chart, err
	}

	if chart.ParentChart.HasParent {
		parentChart := chart.ParentChart.GetParentChartRecursive(s)
		if parentChart == nil {
			return chart, err
		}
		return parentChart, nil
	}
	return chart, nil
}

func (s *ChartStore) getChartFromCache(uri lsp.DocumentURI) *Chart {
	for chartURI, chart := range s.Charts {
		if strings.HasPrefix(uri.Filename(), filepath.Join(chartURI.Filename(), "template")) {
			return chart
		}
	}
	return nil
}

func (s *ChartStore) getChartFromFilesystemForTemplates(path string) (*Chart, error) {
	directory := filepath.Dir(path)
	if filepath.Base(directory) == "templates" {
		templatesDir := directory
		expectedChartDir := filepath.Dir(templatesDir)

		// check if Chart.yaml exists
		if isChartDirectory(expectedChartDir) {
			return s.newChart(uri.File(expectedChartDir), s.valuesFilesConfig), nil
		}
	}

	rootDirectory := s.RootURI.Filename()
	if directory == rootDirectory || directory == path {
		return s.newChart(uri.File(directory), s.valuesFilesConfig), ErrChartNotFound{}
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
	return fmt.Sprintf("Chart not found for file: %s. Using fallback", e.URI)
}
