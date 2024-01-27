package charts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/mrjosh/helm-ls/pkg/chart"
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
	_ = os.WriteFile(chartFile, []byte(chartYaml), 0o644)

	chart := charts.NewChart(uri.File(tempDir), util.ValuesFilesConfig{})
	assert.Equal(t, "hello-world", chart.ChartMetadata.Metadata.Name)
}

func TestNewChartsLoadsDefaultMetadataOnError(t *testing.T) {
	tempDir := t.TempDir()

	chartYaml := `invalidYaml`
	chartFile := filepath.Join(tempDir, "Chart.yaml")
	_ = os.WriteFile(chartFile, []byte(chartYaml), 0o644)

	chart := charts.NewChart(uri.File(tempDir), util.ValuesFilesConfig{})
	assert.Equal(t, "", chart.ChartMetadata.Metadata.Name)
}

func TestNewChartsSetsParentChartURI(t *testing.T) {
	tempDir := t.TempDir()

	chartFile := filepath.Join(tempDir, "Chart.yaml")
	_ = os.WriteFile(chartFile, []byte{}, 0o644)

	chart := charts.NewChart(uri.File(filepath.Join(tempDir, "charts", "subchart")), util.ValuesFilesConfig{})
	assert.Equal(t, tempDir, chart.ParentChart.ParentChartURI.Filename())
}

func TestNewChartsSetsParentChartURIToDefault(t *testing.T) {
	tempDir := t.TempDir()

	chartFile := filepath.Join(tempDir, "Chart.yaml")
	_ = os.WriteFile(chartFile, []byte{}, 0o644)

	chart := charts.NewChart(uri.File(tempDir), util.ValuesFilesConfig{})
	assert.False(t, chart.ParentChart.HasParent)
}

func TestResolvesValuesFileOfParent(t *testing.T) {
	tempDir := t.TempDir()

	chartFile := filepath.Join(tempDir, "Chart.yaml")
	_ = os.WriteFile(chartFile, []byte{}, 0o644)
	valuesFile := filepath.Join(tempDir, "values.yaml")
	_ = os.WriteFile(valuesFile, []byte{}, 0o644)

	subChartValuesFile := filepath.Join(tempDir, "charts", "subchart", "values.yaml")
	subChartChartFile := filepath.Join(tempDir, "charts", "subchart", "Chart.yaml")
	err := os.MkdirAll(filepath.Dir(subChartValuesFile), 0o755)
	assert.NoError(t, err)
	err = os.WriteFile(subChartValuesFile, []byte{}, 0o644)
	assert.NoError(t, err)
	err = os.WriteFile(subChartChartFile, []byte{}, 0o644)
	assert.NoError(t, err)

	chart := charts.NewChart(uri.File(filepath.Join(tempDir, "charts", "subchart")), util.ValuesFilesConfig{})

	expectedChart := &charts.Chart{
		RootURI:       uri.File(tempDir),
		ChartMetadata: &charts.ChartMetadata{},
	}
	newChartFunc := func(_ uri.URI, _ util.ValuesFilesConfig) *charts.Chart { return expectedChart }
	chartStore := charts.NewChartStore(uri.File(tempDir), newChartFunc)

	valueFiles := chart.ResolveValueFiles([]string{"global", "foo"}, chartStore)

	assert.Equal(t, 2, len(valueFiles))
}

func TestResolvesValuesFileOfParentByName(t *testing.T) {
	tempDir := t.TempDir()

	chartFile := filepath.Join(tempDir, "Chart.yaml")
	_ = os.WriteFile(chartFile, []byte{}, 0o644)
	valuesFile := filepath.Join(tempDir, "values.yaml")
	_ = os.WriteFile(valuesFile, []byte{}, 0o644)

	subChartValuesFile := filepath.Join(tempDir, "charts", "subchart", "values.yaml")
	subChartChartFile := filepath.Join(tempDir, "charts", "subchart", "Chart.yaml")
	err := os.MkdirAll(filepath.Dir(subChartValuesFile), 0o755)
	assert.NoError(t, err)
	err = os.WriteFile(subChartValuesFile, []byte{}, 0o644)
	assert.NoError(t, err)
	err = os.WriteFile(subChartChartFile, []byte{}, 0o644)
	assert.NoError(t, err)

	subchart := charts.NewChart(uri.File(filepath.Join(tempDir, "charts", "subchart")), util.ValuesFilesConfig{})
	subchart.ChartMetadata.Metadata.Name = "subchart"

	expectedChart := &charts.Chart{
		RootURI: uri.File(tempDir),
		ChartMetadata: &charts.ChartMetadata{
			Metadata: chart.Metadata{
				Name: "parent",
			},
		},
	}
	newChartFunc := func(_ uri.URI, _ util.ValuesFilesConfig) *charts.Chart { return expectedChart }
	chartStore := charts.NewChartStore(uri.File(tempDir), newChartFunc)

	valueFiles := subchart.ResolveValueFiles([]string{"foo"}, chartStore)

	parentChart, err := chartStore.GetChartForURI(uri.File(tempDir))
	assert.NoError(t, err)

	assert.Equal(t, 2, len(valueFiles))
	assert.Contains(t, valueFiles, &charts.QueriedValuesFiles{Selector: []string{"subchart", "foo"}, ValuesFiles: parentChart.ValuesFiles})
}
