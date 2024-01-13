package charts

import (
	"path/filepath"

	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/uri"
)

type ParentChart struct {
	ParentChartURI uri.URI
	HasParent      bool
}

func getParentChart(rootURI uri.URI) ParentChart {
	directory := filepath.Dir(rootURI.Filename())
	if filepath.Base(directory) == "charts" && isChartDirectory(filepath.Dir(directory)) {
		return ParentChart{uri.New(util.FileURIScheme + filepath.Dir(directory)), true}
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
