package lsp

import (
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

type diagnosticsCache struct {
	YamlDiagnostics []lsp.Diagnostic
	HelmDiagnostics []lsp.Diagnostic
	helmlsConfig    *util.HelmlsConfiguration
}

func NewDiagnosticsCache(helmlsConfig *util.HelmlsConfiguration) diagnosticsCache {
	return diagnosticsCache{
		[]lsp.Diagnostic{},
		[]lsp.Diagnostic{},
		helmlsConfig,
	}
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
	return merged
}
