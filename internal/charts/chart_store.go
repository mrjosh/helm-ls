package charts

import "go.lsp.dev/uri"

type ChartStore struct {
	Charts   map[uri.URI]*Chart
	RootURI  uri.URI
	newChart func(uri.URI) *Chart
}

func NewChartStore(rootURI uri.URI, newChart func(uri.URI) *Chart) *ChartStore {
	return &ChartStore{
		Charts:   map[uri.URI]*Chart{},
		RootURI:  rootURI,
		newChart: newChart,
	}
}

func (s *ChartStore) GetChartForURI(fileURI uri.URI) (*Chart, error) {
	if chart, ok := s.Charts[fileURI]; ok {
		return chart, nil
	}

	var chart *Chart
	expectedChartDir := fileURI.Filename()
	if isChartDirectory(expectedChartDir) {
		chart = s.newChart(uri.New("file://" + expectedChartDir))
	}

	if chart != nil {
		s.Charts[chart.RootURI] = chart
		return chart, nil
	}

	return nil, ErrChartNotFound{
		URI: fileURI,
	}
}
