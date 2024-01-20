package charts

import (
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
	for _, chart := range s.Charts {
		chart.ValuesFiles = NewValuesFiles(chart.RootURI, valuesFilesConfig.MainValuesFileName, valuesFilesConfig.LintOverlayValuesFileName, valuesFilesConfig.AdditionalValuesFilesGlobPattern)
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
