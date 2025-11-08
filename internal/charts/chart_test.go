package charts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/uri"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"

	lsp "go.lsp.dev/protocol"

	"github.com/stretchr/testify/assert"
)

var addChartCallback = func(chart *charts.Chart) {}

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

	sut := charts.NewChart(uri.File(filepath.Join(tempDir, "charts", "subchart")), util.ValuesFilesConfig{})

	expectedChart := &charts.Chart{
		RootURI:       uri.File(tempDir),
		ChartMetadata: &charts.ChartMetadata{},
		HelmChart:     &chart.Chart{},
	}
	newChartFunc := func(_ uri.URI, _ util.ValuesFilesConfig) *charts.Chart { return expectedChart }
	chartStore := charts.NewChartStore(uri.File(tempDir), newChartFunc, addChartCallback)

	valueFiles := sut.ResolveValueFiles([]string{"global", "foo"}, chartStore)

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
		HelmChart: &chart.Chart{},
	}
	newChartFunc := func(_ uri.URI, _ util.ValuesFilesConfig) *charts.Chart { return expectedChart }
	chartStore := charts.NewChartStore(uri.File(tempDir), newChartFunc, addChartCallback)

	valueFiles := subchart.ResolveValueFiles([]string{"foo"}, chartStore)

	parentChart, err := chartStore.GetChartForURI(uri.File(tempDir))
	assert.NoError(t, err)

	assert.Len(t, valueFiles, 2)
	assert.Contains(t, valueFiles, &charts.QueriedValuesFiles{Selector: []string{"subchart", "foo"}, ValuesFiles: parentChart.ValuesFiles})
}

func TestResolvesValuesFileOfDependencyWithGlobal(t *testing.T) {
	var (
		rootDir    = "../../testdata/dependenciesExample"
		chartStore = charts.NewChartStore(uri.File(rootDir), charts.NewChart, addChartCallback)
		chart, err = chartStore.GetChartForDoc(uri.File(filepath.Join(rootDir, "templates", "deployment.yaml")))
		valueFiles = chart.ResolveValueFiles([]string{"global"}, chartStore)
	)

	assert.NoError(t, err)
	assert.Len(t, valueFiles, 3)

	selectors := [][]string{}
	for _, valueFile := range valueFiles {
		selectors = append(selectors, valueFile.Selector)
	}
	assert.Equal(t, selectors, [][]string{{"global"}, {"global"}, {"global"}})
}

func TestResolvesValuesFileOfDependencyWithChartName(t *testing.T) {
	var (
		rootDir    = "../../testdata/dependenciesExample"
		chartStore = charts.NewChartStore(uri.File(rootDir), charts.NewChart, addChartCallback)
		chart, err = chartStore.GetChartForDoc(uri.File(filepath.Join(rootDir, "templates", "deployment.yaml")))
		valueFiles = chart.ResolveValueFiles([]string{"subchartexample", "foo"}, chartStore)
	)

	assert.NoError(t, err)
	assert.Len(t, valueFiles, 2)

	selectors := [][]string{}
	for _, valueFile := range valueFiles {
		selectors = append(selectors, valueFile.Selector)
	}
	assert.Contains(t, selectors, []string{"subchartexample", "foo"})
	assert.Contains(t, selectors, []string{"foo"})
}

func TestResolvesValuesFileOfDependencyWithOnlyChartName(t *testing.T) {
	var (
		rootDir    = "../../testdata/dependenciesExample"
		chartStore = charts.NewChartStore(uri.File(rootDir), charts.NewChart, addChartCallback)
		chart, err = chartStore.GetChartForDoc(uri.File(filepath.Join(rootDir, "templates", "deployment.yaml")))
		valueFiles = chart.ResolveValueFiles([]string{"subchartexample"}, chartStore)
	)

	assert.NoError(t, err)
	assert.Len(t, valueFiles, 2)

	selectors := [][]string{}
	for _, valueFile := range valueFiles {
		selectors = append(selectors, valueFile.Selector)
	}
	assert.Contains(t, selectors, []string{"subchartexample"})
	assert.Contains(t, selectors, []string{})
}

func TestResolvesValuesFileOfDependencyWithChartNameForPackedDependency(t *testing.T) {
	var (
		rootDir    = "../../testdata/dependenciesExample"
		chartStore = charts.NewChartStore(uri.File(rootDir), charts.NewChart, addChartCallback)
		chart, err = chartStore.GetChartForDoc(uri.File(filepath.Join(rootDir, "templates", "deployment.yaml")))
		valueFiles = chart.ResolveValueFiles([]string{"common", "exampleValue"}, chartStore)
	)

	assert.NoError(t, err)
	assert.Len(t, valueFiles, 2)

	selectors := [][]string{}
	for _, valueFile := range valueFiles {
		selectors = append(selectors, valueFile.Selector)
	}
	assert.Contains(t, selectors, []string{"common", "exampleValue"})
	assert.Contains(t, selectors, []string{"exampleValue"})

	var commonValueFile *charts.ValuesFiles
	for _, valueFile := range valueFiles {
		if valueFile.Selector[0] == "exampleValue" {
			commonValueFile = valueFile.ValuesFiles
		}
	}
	assert.NotNil(t, commonValueFile)
	assert.Equal(t, chartutil.Values{"exampleValue": "common-chart"}, commonValueFile.MainValuesFile.Values)
}

func TestLoadsHelmChartWithDependecies(t *testing.T) {
	chart := charts.NewChart(uri.File("../../testdata/dependenciesExample/"), util.ValuesFilesConfig{})

	dependecyTemplates := chart.GetDependeciesTemplates()
	assert.Len(t, dependecyTemplates, 23)

	filePaths := []string{}
	for _, dependency := range dependecyTemplates {
		filePaths = append(filePaths, dependency.Path)
	}
	path, _ := filepath.Abs("../../testdata/dependenciesExample/charts/subchartexample/templates/subchart.yaml")
	assert.Contains(t, filePaths, path)
	path, _ = filepath.Abs("../../testdata/dependenciesExample/charts/" + charts.DependencyCacheFolder + "/common/templates/_names.tpl")
	assert.Contains(t, filePaths, path)
}

func TestGetValueLocation(t *testing.T) {
	chart := charts.NewChart(uri.File("../../testdata/dependenciesExample/"), util.ValuesFilesConfig{})

	valueLocation, err := chart.GetMetadataLocation([]string{"Name"})
	assert.NoError(t, err)

	expected := lsp.Location{
		URI: uri.File("../../testdata/dependenciesExample/Chart.yaml"),
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      1,
				Character: 0,
			},
		},
	}

	assert.Equal(t, expected, valueLocation)
}

func TestLoadsHelmChartWithDependeciesAsFileURI(t *testing.T) {
	chart := charts.NewChart(uri.File("../../testdata/dependencyFileURIExample/"), util.ValuesFilesConfig{})

	dependecyTemplates := chart.GetDependeciesTemplates()
	assert.Len(t, dependecyTemplates, 2)

	filePaths := []string{}
	for _, dependency := range dependecyTemplates {
		filePaths = append(filePaths, dependency.Path)
	}
	path, _ := filepath.Abs("../../testdata/dependenciesExample/charts/subchartexample/templates/subchart.yaml")
	assert.Contains(t, filePaths, path)

	path, _ = filepath.Abs("../../testdata/dependenciesExample/charts/subchartexample/templates/_helpers_subchart.tpl")
	assert.Contains(t, filePaths, path)
}
