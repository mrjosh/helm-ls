package jsonschema

import (
	"encoding/json"
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
	mu             sync.RWMutex
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

func (c *JSONSchemaCache) readCache(uri uri.URI) (cachedGeneratedJSONSchema, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.cache[uri]
	return s, ok
}

func (c *JSONSchemaCache) writeCache(uri uri.URI, cachedGeneratedJSONSchema cachedGeneratedJSONSchema) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[uri] = cachedGeneratedJSONSchema
}

func (c *JSONSchemaCache) GetJSONSchemaForChart(chart *charts.Chart) (string, error) {
	chached, ok := c.readCache(chart.RootURI)

	if !ok {
		return c.createJSONSchemaAndCache(chart)
	}
	if chached.checksum != getChecksum(chart) {
		return c.createJSONSchemaAndCache(chart)
	} else {
		return chached.schemaFilePath, nil
	}
}

func (c *JSONSchemaCache) createJSONSchemaAndCache(chart *charts.Chart) (string, error) {
	logger.Debug("Creating JSON schema for chart", chart.HelmChart.Name())
	generatedChartJSONSchema, err := c.schemaCreation(chart, c.chartStore, c.GetSchemaPathForChart)
	if err != nil {
		logger.Error(err)
		return "", err
	}
	fileName, err := c.writeSchemaToFile(generatedChartJSONSchema.schema, chart)
	if err != nil {
		logger.Error(err)
		return "", err
	}
	c.writeCache(chart.RootURI,
		cachedGeneratedJSONSchema{
			checksum:       getChecksum(chart),
			schemaFilePath: fileName,
		})

	c.processDependencies(generatedChartJSONSchema)

	return fileName, nil
}

func (c *JSONSchemaCache) processDependencies(generatedChartJSONSchema GeneratedChartJSONSchema) {
	var wg sync.WaitGroup

	for _, dependency := range generatedChartJSONSchema.dependencies {
		wg.Add(1)
		// Capture dependency to avoid closure pitfalls (or pass it as a parameter)
		dep := dependency
		go func() {
			defer wg.Done()
			c.GetJSONSchemaForChart(dep)
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

func (c *JSONSchemaCache) writeSchemaToFile(schema *Schema, chart *charts.Chart) (string, error) {
	// TODO: do this only if log level is debug
	bytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}

	path := c.GetSchemaPathForChart(chart)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(bytes); err != nil {
		return "", fmt.Errorf("failed to write schema to file: %w", err)
	}

	return file.Name(), nil
}
