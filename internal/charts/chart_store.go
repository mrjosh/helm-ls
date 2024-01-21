package charts

import (
	"path/filepath"

	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/uri"
)

type ChartStore struct {
	Charts            map[uri.URI]*Chart
	RootURI           uri.URI
	newChart          func(uri.URI, util.ValuesFilesConfig) *Chart
	valuesFilesConfig util.ValuesFilesConfig
}

func NewChartStore(rootURI uri.URI, newChart func(uri.URI, util.ValuesFilesConfig) *Chart) *ChartStore {
	return &ChartStore{
		Charts:            map[uri.URI]*Chart{},
		RootURI:           rootURI,
		newChart:          newChart,
		valuesFilesConfig: util.DefaultConfig.ValuesFilesConfig,
	}
}

func (s *ChartStore) SetValuesFilesConfig(valuesFilesConfig util.ValuesFilesConfig) {
	logger.Debug("SetValuesFilesConfig", valuesFilesConfig)
	if valuesFilesConfig.MainValuesFileName == s.valuesFilesConfig.MainValuesFileName &&
		valuesFilesConfig.AdditionalValuesFilesGlobPattern == s.valuesFilesConfig.AdditionalValuesFilesGlobPattern &&
		valuesFilesConfig.LintOverlayValuesFileName == s.valuesFilesConfig.LintOverlayValuesFileName {
		return
	}
	s.valuesFilesConfig = valuesFilesConfig
	for uri := range s.Charts {
		s.Charts[uri] = s.newChart(uri, valuesFilesConfig)
	}
}

func (s *ChartStore) GetChartForURI(fileURI uri.URI) (*Chart, error) {
	if chart, ok := s.Charts[fileURI]; ok {
		return chart, nil
	}

	var chart *Chart
	expectedChartDir := fileURI.Filename()
	if isChartDirectory(expectedChartDir) {
		chart = s.newChart(uri.New("file://"+expectedChartDir), s.valuesFilesConfig)
	}

	if chart != nil {
		s.Charts[chart.RootURI] = chart
		return chart, nil
	}

	return nil, ErrChartNotFound{
		URI: fileURI,
	}
}

func (s *ChartStore) ReloadValuesFile(file uri.URI) {
	logger.Println("Reloading values file", file)
	chart, err := s.GetChartForURI(uri.URI(util.FileURIScheme + filepath.Dir(file.Filename())))
	if err != nil {
		logger.Error("Error reloading values file", file, err)
		return
	}
	for _, valuesFile := range chart.ValuesFiles.AllValuesFiles() {
		if valuesFile.URI == file {
			valuesFile.Reload()
		}
	}
}
