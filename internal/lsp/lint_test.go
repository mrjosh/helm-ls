package lsp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
	"helm.sh/helm/v3/pkg/chartutil"
)

func TestLint(t *testing.T) {
	diagnostics := GetDiagnostics(uri.File("../../testdata/example/templates/lint.yaml"), chartutil.Values{})
	assert.NotEmpty(t, diagnostics)
}
