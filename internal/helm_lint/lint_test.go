package helmlint

import (
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
	"helm.sh/helm/v3/pkg/chartutil"
)

func TestLint(t *testing.T) {
	diagnostics := GetDiagnostics(uri.File("../../testdata/example"), chartutil.Values{})
	assert.NotEmpty(t, diagnostics)
	assert.Len(t, diagnostics, 2)
	assert.Len(t, diagnostics[uri.File("../../testdata/example/Chart.yaml").Filename()], 1)
}

func TestLintNotifications(t *testing.T) {
	chart := charts.Chart{
		RootURI: uri.File("../../testdata/example"),
		ValuesFiles: &charts.ValuesFiles{
			MainValuesFile:        &charts.ValuesFile{},
			OverlayValuesFile:     &charts.ValuesFile{},
			AdditionalValuesFiles: []*charts.ValuesFile{},
		},
	}
	diagnostics := GetDiagnosticsNotifications(&chart, &document.TemplateDocument{
		Document: document.Document{
			URI: uri.File("../../testdata/example/templates/deployment-no-templates.yaml"),
		},
	})
	assert.NotEmpty(t, diagnostics)
	assert.Len(t, diagnostics, 3)

	uris := []string{}
	for _, notification := range diagnostics {
		uris = append(uris, notification.URI.Filename())
	}
	assert.Contains(t, uris, uri.File("../../testdata/example/templates/deployment-no-templates.yaml").Filename())
	for _, notification := range diagnostics {
		if notification.URI.Filename() == uri.File("../../testdata/example/templates/deployment-no-templates.yaml").Filename() {
			assert.Empty(t, notification.Diagnostics)
		}
	}
}

func TestLintNotificationsIncludesEmptyDiagnosticsForFixedIssues(t *testing.T) {
	chart := charts.Chart{
		RootURI: uri.File("../../testdata/example"),
		ValuesFiles: &charts.ValuesFiles{
			MainValuesFile:        &charts.ValuesFile{},
			OverlayValuesFile:     &charts.ValuesFile{},
			AdditionalValuesFiles: []*charts.ValuesFile{},
		},
	}
	diagnostics := GetDiagnosticsNotifications(&chart, &document.TemplateDocument{
		Document: document.Document{URI: uri.File("../../testdata/example/templates/deployment-no-templates.yaml")},
	},
	)

	uris := []string{}
	for _, notification := range diagnostics {
		uris = append(uris, notification.URI.Filename())
	}
	assert.Contains(t, uris, uri.File("../../testdata/example/templates/deployment-no-templates.yaml").Filename())
	for _, notification := range diagnostics {
		if notification.URI.Filename() == uri.File("../../testdata/example/templates/deployment-no-templates.yaml").Filename() {
			assert.Empty(t, notification.Diagnostics)
		}
	}
}
