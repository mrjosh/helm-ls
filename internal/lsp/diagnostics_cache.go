package lsp

import (
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

type diagnosticsCache struct {
	YamlDiagnostics             []lsp.Diagnostic
	HelmDiagnostics             []lsp.Diagnostic
	helmlsConfig                util.HelmlsConfiguration
	gotYamlDiagnosticsTimes     int
	yamlDiagnosticsCountReduced bool
}

func NewDiagnosticsCache(helmlsConfig util.HelmlsConfiguration) diagnosticsCache {
	return diagnosticsCache{
		[]lsp.Diagnostic{},
		[]lsp.Diagnostic{},
		helmlsConfig,
		0,
		false,
	}
}

func (d *diagnosticsCache) SetYamlDiagnostics(diagnostics []lsp.Diagnostic) {
	d.yamlDiagnosticsCountReduced = len(diagnostics) < len(d.YamlDiagnostics)
	d.YamlDiagnostics = diagnostics
	d.gotYamlDiagnosticsTimes++
}

func (d diagnosticsCache) GetMergedDiagnostics() (merged []lsp.Diagnostic) {
	merged = []lsp.Diagnostic{}
	for _, diagnostic := range d.HelmDiagnostics {
		merged = append(merged, diagnostic)
	}
	for i, diagnostic := range d.YamlDiagnostics {
		if i < d.helmlsConfig.YamllsConfiguration.DiagnosticsLimit {
			merged = append(merged, diagnostic)
		}
	}
	logger.Debug("Merged diagnostics", merged)
	return merged
}

func (d *diagnosticsCache) ShouldShowDiagnosticsOnNewYamlDiagnostics() bool {

	return d.yamlDiagnosticsCountReduced || // show the diagnostics when the count is reduced, this means an error was fixed and it should be shown to the user
		d.helmlsConfig.YamllsConfiguration.ShowDiagnosticsDirectly || // show the diagnostics directly when the user configured to show them
		d.gotYamlDiagnosticsTimes < 3 // show the diagnostics, when it are the inital diagnostics that are sent after opening a file. Initial diagnostics are sent twice from yamlls
}
