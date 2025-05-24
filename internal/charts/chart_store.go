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
	addChartCallback  func(chart *Chart)
	valuesFilesConfig util.ValuesFilesConfig
}

func NewChartStore(rootURI uri.URI, newChart func(uri.URI, util.ValuesFilesConfig) *Chart, addChartCallback func(chart *Chart)) *ChartStore {
	return &ChartStore{
		Charts:            map[uri.URI]*Chart{},
		RootURI:           rootURI,
		newChart:          newChart,
		addChartCallback:  addChartCallback,
		valuesFilesConfig: util.DefaultConfig.ValuesFilesConfig,
	}
}

func (s *ChartStore) SetRootURI(rootURI uri.URI) {
	s.RootURI = rootURI
	var uris []uri.URI
	for chartURI := range s.Charts {
		uris = append(uris, chartURI)
	}
	for _, uri := range uris {
		s.AddChart(s.newChart(uri, s.valuesFilesConfig))
	}
}

// AddChart adds a new chart to the store and loads its dependencies
func (s *ChartStore) AddChart(chart *Chart) {
	s.Charts[chart.RootURI] = chart
	s.loadChartDependencies(chart)
	s.addChartCallback(chart)
}

func (s *ChartStore) SetValuesFilesConfig(valuesFilesConfig util.ValuesFilesConfig) {
	logger.Debug("SetValuesFilesConfig", valuesFilesConfig)
	if valuesFilesConfig.MainValuesFileName == s.valuesFilesConfig.MainValuesFileName &&
		valuesFilesConfig.AdditionalValuesFilesGlobPattern == s.valuesFilesConfig.AdditionalValuesFilesGlobPattern &&
		valuesFilesConfig.LintOverlayValuesFileName == s.valuesFilesConfig.LintOverlayValuesFileName {
		return
	}
	s.valuesFilesConfig = valuesFilesConfig
	var uris []uri.URI
	for chartURI := range s.Charts {
		uris = append(uris, chartURI)
	}
	for _, chartURI := range uris {
		s.AddChart(s.newChart(chartURI, valuesFilesConfig))
	}
}

func (s *ChartStore) GetChartForURI(fileURI uri.URI) (*Chart, error) {
	if chart, ok := s.Charts[fileURI]; ok {
		return chart, nil
	}

	var chart *Chart
	expectedChartDir := fileURI.Filename()
	if isChartDirectory(expectedChartDir) {
		chart = s.newChart(uri.File(expectedChartDir), s.valuesFilesConfig)
	}

	if chart != nil {
		s.AddChart(chart)
		return chart, nil
	}

	return nil, ErrChartNotFound{
		URI: fileURI,
	}
}

func (s *ChartStore) ReloadValuesFile(file uri.URI) {
	logger.Println("Reloading values file", file)
	chart, err := s.GetChartForURI(uri.File(filepath.Dir(file.Filename())))
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

func (s *ChartStore) loadChartDependencies(chart *Chart) {
	for _, dependency := range chart.HelmChart.Dependencies() {
		dependencyURI := chart.GetDependecyURI(dependency.Name())
		chart := NewChartFromHelmChart(dependency, dependencyURI)

		s.AddChart(chart)
	}
}
