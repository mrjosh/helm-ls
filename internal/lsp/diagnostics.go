package lsp

import lsp "go.lsp.dev/protocol"

type diagnosticsCache struct {
	YamlDiagnostics []lsp.Diagnostic
	HelmDiagnostics []lsp.Diagnostic
}

// TODO: this should be configurable
// max diagnostics that are shown for a single file
const yamlDiagnosticsLimit = 50

func NewDiagnosticsCache() diagnosticsCache {
	return diagnosticsCache{
		[]lsp.Diagnostic{},
		[]lsp.Diagnostic{},
	}
}

func (d diagnosticsCache) GetMergedDiagnostics() (merged []lsp.Diagnostic) {
	merged = []lsp.Diagnostic{}
	for _, diagnostic := range d.HelmDiagnostics {
		merged = append(merged, diagnostic)
	}
	for i, diagnostic := range d.YamlDiagnostics {
		if i < yamlDiagnosticsLimit {
			merged = append(merged, diagnostic)
		}
	}
	return merged
}
