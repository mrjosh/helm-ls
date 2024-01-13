package charts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"go.lsp.dev/uri"

	"github.com/stretchr/testify/assert"
)

func TestNewChartsLoadsMetadata(t *testing.T) {
	tempDir := t.TempDir()

	chartYaml := `
apiVersion: v2
name: hello-world
description: A Helm chart for Kubernetes
type: application`
	chartFile := filepath.Join(tempDir, "Chart.yaml")
	_ = os.WriteFile(chartFile, []byte(chartYaml), 0644)

	chart := charts.NewChart(uri.New("file://" + tempDir))
	assert.Equal(t, "hello-world", chart.ChartMetadata.Metadata.Name)
}

func TestNewChartsLoadsDefaultMetadataOnError(t *testing.T) {
	tempDir := t.TempDir()

	chartYaml := `invalidYaml`
	chartFile := filepath.Join(tempDir, "Chart.yaml")
	_ = os.WriteFile(chartFile, []byte(chartYaml), 0644)

	chart := charts.NewChart(uri.New("file://" + tempDir))
	assert.Equal(t, "", chart.ChartMetadata.Metadata.Name)
}

func TestNewChartsSetsParentChartURI(t *testing.T) {
	tempDir := t.TempDir()

	chartFile := filepath.Join(tempDir, "Chart.yaml")
	_ = os.WriteFile(chartFile, []byte{}, 0644)

	chart := charts.NewChart(uri.New("file://" + filepath.Join(tempDir, "charts", "subchart")))
	assert.Equal(t, tempDir, chart.ParentChart.ParentChartURI.Filename())
}

func TestNewChartsSetsParentChartURIToDefault(t *testing.T) {
	tempDir := t.TempDir()

	chartFile := filepath.Join(tempDir, "Chart.yaml")
	_ = os.WriteFile(chartFile, []byte{}, 0644)

	chart := charts.NewChart(uri.New("file://" + tempDir))
	assert.False(t, chart.ParentChart.HasParent)
}

func TestResolvesValuesFileOfParent(t *testing.T) {

	tempDir := t.TempDir()

	chartFile := filepath.Join(tempDir, "Chart.yaml")
	_ = os.WriteFile(chartFile, []byte{}, 0644)
	valuesFile := filepath.Join(tempDir, "values.yaml")
	_ = os.WriteFile(valuesFile, []byte{}, 0644)

	subChartValuesFile := filepath.Join(tempDir, "charts", "subchart", "values.yaml")
	subChartChartFile := filepath.Join(tempDir, "charts", "subchart", "Chart.yaml")
	var err = os.MkdirAll(filepath.Dir(subChartValuesFile), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(subChartValuesFile, []byte{}, 0644)
	assert.NoError(t, err)
	err = os.WriteFile(subChartChartFile, []byte{}, 0644)
	assert.NoError(t, err)

	chart := charts.NewChart(uri.New("file://" + filepath.Join(tempDir, "charts", "subchart")))

	expectedChart := &charts.Chart{
		RootURI: uri.New("file://" + tempDir),
	}
	newChartFunc := func(_ uri.URI) *charts.Chart { return expectedChart }
	chartStore := charts.NewChartStore(uri.New("file://"+tempDir), newChartFunc)

	valueFiles := chart.ResolveValueFiles([]string{"global", "foo"}, chartStore)

	assert.Equal(t, 2, len(valueFiles))

}
