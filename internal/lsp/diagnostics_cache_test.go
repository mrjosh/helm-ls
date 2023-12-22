package lsp

import (
	"github.com/mrjosh/helm-ls/internal/util"

	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"
	"testing"
)

func TestDiagnosticsCache_SetYamlDiagnostics(t *testing.T) {
	helmlsConfig := util.HelmlsConfiguration{}
	cache := NewDiagnosticsCache(helmlsConfig)

	diagnostics := []lsp.Diagnostic{
		{
			Range:              lsp.Range{},
			Severity:           0,
			Code:               nil,
			CodeDescription:    &lsp.CodeDescription{},
			Source:             "",
			Message:            "test",
			Tags:               []lsp.DiagnosticTag{},
			RelatedInformation: []lsp.DiagnosticRelatedInformation{},
			Data:               nil,
		},
	}

	cache.SetYamlDiagnostics(diagnostics)

	assert.Equal(t, diagnostics, cache.YamlDiagnostics)
	assert.False(t, cache.yamlDiagnosticsCountReduced)
	assert.Equal(t, 1, cache.gotYamlDiagnosticsTimes)
}

func TestDiagnosticsCache_GetMergedDiagnostics(t *testing.T) {
	helmlsConfig := util.DefaultConfig
	cache := NewDiagnosticsCache(helmlsConfig)

	yamlDiagnostics := []lsp.Diagnostic{
		{
			Range:              lsp.Range{},
			Severity:           0,
			Code:               nil,
			CodeDescription:    &lsp.CodeDescription{},
			Source:             "",
			Message:            "test",
			Tags:               []lsp.DiagnosticTag{},
			RelatedInformation: []lsp.DiagnosticRelatedInformation{},
			Data:               nil,
		},
	}
	helmDiagnostics := []lsp.Diagnostic{
		{
			Range:              lsp.Range{},
			Severity:           0,
			Code:               nil,
			CodeDescription:    &lsp.CodeDescription{},
			Source:             "",
			Message:            "test2",
			Tags:               []lsp.DiagnosticTag{},
			RelatedInformation: []lsp.DiagnosticRelatedInformation{},
			Data:               nil,
		},
	}

	cache.SetYamlDiagnostics(yamlDiagnostics)
	cache.HelmDiagnostics = helmDiagnostics
	merged := cache.GetMergedDiagnostics()

	assert.Equal(t, 2, len(merged))
}

func TestDiagnosticsCache_ShouldShowDiagnosticsOnNewYamlDiagnostics(t *testing.T) {
	helmlsConfig := util.HelmlsConfiguration{}
	cache := NewDiagnosticsCache(helmlsConfig)
	yamlDiagnostics := []lsp.Diagnostic{
		{
			Range:              lsp.Range{},
			Severity:           0,
			Code:               nil,
			CodeDescription:    &lsp.CodeDescription{},
			Source:             "",
			Message:            "test",
			Tags:               []lsp.DiagnosticTag{},
			RelatedInformation: []lsp.DiagnosticRelatedInformation{},
			Data:               nil,
		},
	}
	cache.SetYamlDiagnostics(yamlDiagnostics)
	cache.SetYamlDiagnostics([]lsp.Diagnostic{})

	if !cache.ShouldShowDiagnosticsOnNewYamlDiagnostics() {
		t.Error("Expected ShouldShowDiagnosticsOnNewYamlDiagnostics to be true, but got false")
	}

	cache.SetYamlDiagnostics(yamlDiagnostics)

	if cache.ShouldShowDiagnosticsOnNewYamlDiagnostics() {
		t.Error("Expected ShouldShowDiagnosticsOnNewYamlDiagnostics to be false, but got true")
	}
}

func TestDiagnosticsCache_ShouldShowDiagnosticsOnNewYamlDiagnosticsForInitialDiagnostics(t *testing.T) {
	helmlsConfig := util.HelmlsConfiguration{}
	cache := NewDiagnosticsCache(helmlsConfig)
	yamlDiagnostics := []lsp.Diagnostic{
		{
			Range:              lsp.Range{},
			Severity:           0,
			Code:               nil,
			CodeDescription:    &lsp.CodeDescription{},
			Source:             "",
			Message:            "test",
			Tags:               []lsp.DiagnosticTag{},
			RelatedInformation: []lsp.DiagnosticRelatedInformation{},
			Data:               nil,
		},
	}
	cache.SetYamlDiagnostics(yamlDiagnostics)
	cache.SetYamlDiagnostics(yamlDiagnostics)

	assert.True(t, cache.ShouldShowDiagnosticsOnNewYamlDiagnostics())

	cache.SetYamlDiagnostics(yamlDiagnostics)

	assert.False(t, cache.ShouldShowDiagnosticsOnNewYamlDiagnostics())
}
