package charts

import (
	"path/filepath"

	"go.lsp.dev/uri"
)

type ParentChart struct {
	ParentChartURI uri.URI
	HasParent      bool
}

func newParentChart(rootURI uri.URI) ParentChart {
	directory := filepath.Dir(rootURI.Filename())
	if filepath.Base(directory) == "charts" && isChartDirectory(filepath.Dir(directory)) {
		return ParentChart{uri.File(filepath.Dir(directory)), true}
	}
	return ParentChart{}
}

func (p *ParentChart) GetParentChart(chartStore *ChartStore) *Chart {
	if !p.HasParent {
		return nil
	}
	chart, err := chartStore.GetChartForURI(p.ParentChartURI)
	if err != nil {
		logger.Error("Error getting parent chart ", err)
		return nil
	}
	return chart
}

func (p *ParentChart) GetParentChartRecursive(chartStore *ChartStore) *Chart {
	chart := p.GetParentChart(chartStore)
	if chart == nil {
		return nil
	}
	parentChart := chart.ParentChart.GetParentChartRecursive(chartStore)
	if parentChart == nil {
		return chart
	}
	return parentChart
}
