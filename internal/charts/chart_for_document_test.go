package charts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
)

func TestGetChartForDocumentWorksForAlreadyAddedCharts(t *testing.T) {
	var chartStore = charts.NewChartStore("file:///tmp", func(_ uri.URI) *charts.Chart {
		return &charts.Chart{}
	})

	chart := &charts.Chart{}
	chartStore.Charts["file:///tmp/chart"] = chart
	subchart := &charts.Chart{}
	chartStore.Charts["file:///tmp/chart/charts/subchart"] = subchart
	otherchart := &charts.Chart{}
	chartStore.Charts["file:///tmp/otherChart"] = otherchart

	result1, error := chartStore.GetChartForDoc("file:///tmp/chart/templates/deployment.yaml")
	assert.NoError(t, error)
	assert.Same(t, chart, result1)

	result2, error := chartStore.GetChartForDoc("file:///tmp/chart/templates/directory/deployment.yaml")
	assert.NoError(t, error)
	assert.Same(t, chart, result2)

	result3, error := chartStore.GetChartForDoc("file:///tmp/chart/charts/subchart/templates/directory/deployment.yaml")
	assert.NoError(t, error)
	assert.Same(t, subchart, result3)

	result4, error := chartStore.GetChartForDoc("file:///tmp/otherChart/templates/directory/deployment.yaml")
	assert.NoError(t, error)
	assert.Same(t, otherchart, result4)

	result5, error := chartStore.GetChartForDoc("file:///tmp/directory/deployment.yaml")
	assert.Error(t, error)
	assert.Nil(t, result5)
}

func TestGetChartForDocumentWorksForNewToAddChart(t *testing.T) {
	var (
		rootDir                = t.TempDir()
		expectedChartDirectory = filepath.Join(rootDir, "chart")
		expectedChart          = &charts.Chart{
			RootURI: uri.New("file://" + expectedChartDirectory),
		}
		newChartFunc = func(_ uri.URI) *charts.Chart { return expectedChart }
		chartStore   = charts.NewChartStore(uri.New("file://"+rootDir), newChartFunc)
		err          = os.MkdirAll(expectedChartDirectory, 0755)
	)
	assert.NoError(t, err)
	_, _ = os.Create(filepath.Join(expectedChartDirectory, "Chart.yaml"))

	result1, error := chartStore.GetChartForDoc(uri.New("file://" + filepath.Join(expectedChartDirectory, "templates", "deployment.yaml")))
	assert.NoError(t, error)
	assert.Same(t, expectedChart, result1)

	assert.Same(t, expectedChart, chartStore.Charts[uri.New("file://"+expectedChartDirectory)])
}

func TestGetChartForDocumentWorksForNewToAddChartWithNestedFile(t *testing.T) {
	var (
		rootDir                = t.TempDir()
		expectedChartDirectory = filepath.Join(rootDir, "chart")
		expectedChart          = &charts.Chart{
			RootURI: uri.New("file://" + expectedChartDirectory),
		}
		newChartFunc = func(_ uri.URI) *charts.Chart { return expectedChart }
		chartStore   = charts.NewChartStore(uri.New("file://"+rootDir), newChartFunc)
		err          = os.MkdirAll(expectedChartDirectory, 0755)
	)
	assert.NoError(t, err)
	_, _ = os.Create(filepath.Join(expectedChartDirectory, "Chart.yaml"))

	result1, error := chartStore.GetChartForDoc(uri.New("file://" + filepath.Join(expectedChartDirectory, "templates", "nested", "deployment.yaml")))
	assert.NoError(t, error)
	assert.Same(t, expectedChart, result1)

	assert.Same(t, expectedChart, chartStore.Charts[uri.New("file://"+expectedChartDirectory)])
}
