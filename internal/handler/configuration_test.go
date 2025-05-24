package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/util"
	mocks "github.com/mrjosh/helm-ls/mocks/go.lsp.dev/protocol"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

var (
	addChartCallback    = func(chart *charts.Chart) {}
	configurationParams = lsp.ConfigurationParams{Items: []lsp.ConfigurationItem{{Section: "helm-ls"}}}
)

func TestConfigurationWorks(t *testing.T) {
	mockClient := mocks.NewMockClient(t)
	handler := &ServerHandler{
		helmlsConfig: util.DefaultConfig,
		chartStore:   charts.NewChartStore(uri.File("/"), charts.NewChart, addChartCallback),
	}
	handler.client = mockClient

	userConfig := []any{map[string]any{
		"LogLevel": "debug",
		// disable yamlls to avoid configuring it in the test
		"yamlls": map[string]any{"enabled": false},
	}}
	mockClient.EXPECT().Configuration(mock.Anything, &configurationParams).Return(userConfig, nil)
	handler.retrieveWorkspaceConfiguration(context.Background())

	expectedConfig := util.DefaultConfig
	expectedConfig.LogLevel = "debug"
	expectedConfig.YamllsConfiguration.Enabled = false
	assert.Equal(t, expectedConfig, handler.helmlsConfig)
}

func TestConfigurationWorksForEmptyConfig(t *testing.T) {
	mockClient := mocks.NewMockClient(t)
	handler := &ServerHandler{
		helmlsConfig: util.DefaultConfig,
		chartStore:   charts.NewChartStore(uri.File("/"), charts.NewChart, addChartCallback),
	}
	handler.client = mockClient
	// disable yamlls to avoid configuring it in the test
	handler.helmlsConfig.YamllsConfiguration.Enabled = false

	userConfig := []any{}
	mockClient.EXPECT().Configuration(mock.Anything, &configurationParams).Return(userConfig, nil)
	handler.retrieveWorkspaceConfiguration(context.Background())

	expectedConfig := util.DefaultConfig
	expectedConfig.YamllsConfiguration.Enabled = false
	assert.Equal(t, expectedConfig, handler.helmlsConfig)
}

func TestConfigurationWorksForError(t *testing.T) {
	mockClient := mocks.NewMockClient(t)
	handler := &ServerHandler{
		helmlsConfig: util.DefaultConfig,
		chartStore:   charts.NewChartStore(uri.File("/"), charts.NewChart, addChartCallback),
	}
	handler.client = mockClient

	// disable yamlls to avoid configuring it in the test
	handler.helmlsConfig.YamllsConfiguration.Enabled = false

	userConfig := []any{map[string]any{
		"LogLevel": "debug",
	}}
	mockClient.EXPECT().Configuration(mock.Anything, &configurationParams).Return(userConfig, errors.New("error"))
	handler.retrieveWorkspaceConfiguration(context.Background())

	expectedConfig := util.DefaultConfig
	expectedConfig.YamllsConfiguration.Enabled = false
	assert.Equal(t, expectedConfig, handler.helmlsConfig)
}

func TestConfigurationWorksForJsonError(t *testing.T) {
	mockClient := mocks.NewMockClient(t)
	handler := &ServerHandler{
		helmlsConfig: util.DefaultConfig,
		chartStore:   charts.NewChartStore(uri.File("/"), charts.NewChart, addChartCallback),
	}
	handler.client = mockClient

	// disable yamlls to avoid configuring it in the test
	handler.helmlsConfig.YamllsConfiguration.Enabled = false

	userConfig := []any{map[string]any{
		"LogLevel": "debug",
		"test":     func() {},
	}}
	mockClient.EXPECT().Configuration(mock.Anything, &configurationParams).Return(userConfig, nil)
	handler.retrieveWorkspaceConfiguration(context.Background())

	expectedConfig := util.DefaultConfig
	expectedConfig.YamllsConfiguration.Enabled = false
	assert.Equal(t, expectedConfig, handler.helmlsConfig)
}

func TestConfigurationWorksForYamllsGlob(t *testing.T) {
	mockClient := mocks.NewMockClient(t)
	handler := &ServerHandler{
		helmlsConfig: util.DefaultConfig,
		chartStore:   charts.NewChartStore(uri.File("/"), charts.NewChart, addChartCallback),
	}
	handler.client = mockClient

	userConfig := []any{map[string]any{
		// disable yamlls to avoid configuring it in the test
		"yamlls": map[string]any{"enabled": false, "enabledForFilesGlob": "*.ext"},
	}}
	mockClient.EXPECT().Configuration(mock.Anything, &configurationParams).Return(userConfig, nil)
	handler.retrieveWorkspaceConfiguration(context.Background())

	assert.Equal(t, handler.helmlsConfig.YamllsConfiguration.Enabled, false)
	assert.Equal(t, handler.helmlsConfig.YamllsConfiguration.EnabledForFilesGlob, "*.ext")

	assert.True(t, handler.helmlsConfig.YamllsConfiguration.EnabledForFilesGlobObject.Match("file.ext"))
	assert.False(t, handler.helmlsConfig.YamllsConfiguration.EnabledForFilesGlobObject.Match("file.yaml"))
}
