package lsp

import lsp "go.lsp.dev/protocol"

type diagnosticsCache struct {
	Yamldiagnostics []lsp.Diagnostic
	Helmdiagnostics []lsp.Diagnostic
}

func NewDiagnosticsCache() diagnosticsCache {
	return diagnosticsCache{
		[]lsp.Diagnostic{},
		[]lsp.Diagnostic{},
	}
}

func (d diagnosticsCache) GetMergedDiagnostics() (merged []lsp.Diagnostic) {
	merged = []lsp.Diagnostic{}
	for _, diagnostic := range d.Yamldiagnostics {
		merged = append(merged, diagnostic)
	}
	for _, diagnostic := range d.Helmdiagnostics {
		merged = append(merged, diagnostic)
	}
	return merged
}
