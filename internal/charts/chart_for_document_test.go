package charts_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
)

func TestGetChartForDocumentWorksForAlreadyAddedCharts(t *testing.T) {
	chartStore := charts.NewChartStore("file:///tmp", func(uri uri.URI, _ util.ValuesFilesConfig) *charts.Chart {
		return &charts.Chart{RootURI: uri}
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
	assert.Equal(t, &charts.Chart{RootURI: uri.File("/tmp")}, result5)
}

func TestGetChartForDocumentWorksForNewToAddChart(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on windows because of https://github.com/golang/go/issues/51442")
	}
	var (
		rootDir                = t.TempDir()
		expectedChartDirectory = filepath.Join(rootDir, "chart")
		expectedChart          = &charts.Chart{
			RootURI: uri.File(expectedChartDirectory),
		}
		newChartFunc = func(_ uri.URI, _ util.ValuesFilesConfig) *charts.Chart { return expectedChart }
		chartStore   = charts.NewChartStore(uri.File(rootDir), newChartFunc)
		err          = os.MkdirAll(expectedChartDirectory, 0o755)
	)
	assert.NoError(t, err)
	chartFile := filepath.Join(expectedChartDirectory, "Chart.yaml")
	_, _ = os.Create(chartFile)

	result1, error := chartStore.GetChartForDoc(uri.File(filepath.Join(expectedChartDirectory, "templates", "deployment.yaml")))
	assert.NoError(t, error)
	assert.Same(t, expectedChart, result1)

	assert.Same(t, expectedChart, chartStore.Charts[uri.File(expectedChartDirectory)])
}

func TestGetChartForDocumentWorksForNewToAddChartWithNestedFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on windows because of https://github.com/golang/go/issues/51442")
	}
	var (
		rootDir                = t.TempDir()
		expectedChartDirectory = filepath.Join(rootDir, "chart")
		expectedChart          = &charts.Chart{
			RootURI: uri.File(expectedChartDirectory),
		}
		newChartFunc = func(_ uri.URI, _ util.ValuesFilesConfig) *charts.Chart { return expectedChart }
		chartStore   = charts.NewChartStore(uri.File(rootDir), newChartFunc)
		err          = os.MkdirAll(expectedChartDirectory, 0o755)
	)
	assert.NoError(t, err)
	chartFile := filepath.Join(expectedChartDirectory, "Chart.yaml")
	_, _ = os.Create(chartFile)

	result1, error := chartStore.GetChartForDoc(uri.File(filepath.Join(expectedChartDirectory, "templates", "nested", "deployment.yaml")))
	assert.NoError(t, error)
	assert.Same(t, expectedChart, result1)

	assert.Same(t, expectedChart, chartStore.Charts[uri.File(expectedChartDirectory)])
}
