package jsonschema

import (
	"hash/adler32"

	"github.com/mrjosh/helm-ls/internal/charts"
	"go.lsp.dev/uri"
)

type cachedGeneratedJSONSchema struct {
	checksum       uint32
	schemaFilePath string
}

type JSONSchemaCache struct {
	cache          map[uri.URI]cachedGeneratedJSONSchema
	schemaCreation func(chart *charts.Chart, chartStore *charts.ChartStore) (string, error)
	chartStore     *charts.ChartStore
}

func NewJSONSchemaCache(chartStore *charts.ChartStore) *JSONSchemaCache {
	return &JSONSchemaCache{
		cache:          make(map[uri.URI]cachedGeneratedJSONSchema),
		schemaCreation: CreateJsonSchemaForChart,
		chartStore:     chartStore,
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
	schemaFilePath, err := c.schemaCreation(chart, c.chartStore)
	if err != nil {
		logger.Error(err)
		return "", err
	}
	c.cache[chart.RootURI] = cachedGeneratedJSONSchema{
		checksum:       getChecksum(chart),
		schemaFilePath: schemaFilePath,
	}
	return schemaFilePath, nil
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
