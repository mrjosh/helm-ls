package jsonschema

import (
	"fmt"
	"hash/adler32"
	"os"
	"path/filepath"
	"sync"

	"github.com/mrjosh/helm-ls/internal/charts"
	"go.lsp.dev/uri"
)

type cachedGeneratedJSONSchema struct {
	checksum       uint32
	schemaFilePath string
}

type JSONSchemaCache struct {
	cache          map[uri.URI]cachedGeneratedJSONSchema
	schemaCreation func(chart *charts.Chart, chartStore *charts.ChartStore, getSchemaPathForChart func(chart *charts.Chart) string) (GeneratedChartJSONSchema, error)
	chartStore     *charts.ChartStore
	schemaFilesDir string
}

func NewJSONSchemaCache(chartStore *charts.ChartStore) *JSONSchemaCache {
	schemaFilesDir := filepath.Join(os.TempDir(), "helm-ls")

	err := os.MkdirAll(schemaFilesDir, os.ModePerm)
	if err != nil {
		logger.Error("Failed to create schema files directory:", err)
	}

	return &JSONSchemaCache{
		cache:          make(map[uri.URI]cachedGeneratedJSONSchema),
		schemaCreation: CreateJsonSchemaForChart,
		chartStore:     chartStore,
		schemaFilesDir: schemaFilesDir,
	}
}

func (c *JSONSchemaCache) GetJsonSchemaForChart(chart *charts.Chart) (string, error) {
	chached, ok := c.cache[chart.RootURI]

	if !ok {
		return c.createJsonSchemaAndCache(chart)
	}
	if chached.checksum != getChecksum(chart) {
		return c.createJsonSchemaAndCache(chart)
	} else {
		return chached.schemaFilePath, nil
	}
}

func (c *JSONSchemaCache) createJsonSchemaAndCache(chart *charts.Chart) (string, error) {
	logger.Println("Creating JSON schema for chart", chart.HelmChart.Name())
	generatedChartJSONSchema, err := c.schemaCreation(chart, c.chartStore, c.GetSchemaPathForChart)
	if err != nil {
		logger.Error(err)
		return "", err
	}
	c.cache[chart.RootURI] = cachedGeneratedJSONSchema{
		checksum:       getChecksum(chart),
		schemaFilePath: generatedChartJSONSchema.path,
	}

	// c.processDependencies(generatedChartJSONSchema)
	for _, dependency := range generatedChartJSONSchema.dependencies {
		// IDEA: parallel this
		c.GetJsonSchemaForChart(dependency)
	}

	return generatedChartJSONSchema.path, nil
}

func (c *JSONSchemaCache) processDependencies(generatedChartJSONSchema GeneratedChartJSONSchema) {
	var wg sync.WaitGroup

	for _, dependency := range generatedChartJSONSchema.dependencies {
		wg.Add(1)
		// Capture dependency to avoid closure pitfalls (or pass it as a parameter)
		dep := dependency
		go func() {
			defer wg.Done()
			c.GetJsonSchemaForChart(dep)
		}()
	}

	wg.Wait()
}

func getChecksum(chart *charts.Chart) uint32 {
	totalContent := []byte{}

	for _, value := range chart.ValuesFiles.AllValuesFiles() {
		content, err := value.Values.YAML()
		if err != nil {
			logger.Error(err)
			continue
		}
		totalContent = append(totalContent, content...)
	}

	return adler32.Checksum(totalContent)
}

func (c *JSONSchemaCache) GetSchemaPathForChart(chart *charts.Chart) string {
	id := getChecksum(chart)

	return filepath.Join(c.schemaFilesDir, fmt.Sprintf("%d-%s.json", id, chart.HelmChart.Name()))
}
