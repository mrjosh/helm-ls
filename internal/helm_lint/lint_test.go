package helmlint

import (
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
	"helm.sh/helm/v3/pkg/chartutil"
)

func TestLint(t *testing.T) {
	diagnostics := GetDiagnostics(uri.File("../../testdata/example"), chartutil.Values{}, []string{})
	assert.NotEmpty(t, diagnostics)
	assert.Len(t, diagnostics, 2)
	assert.Len(t, diagnostics[uri.File("../../testdata/example/Chart.yaml").Filename()], 1)
}

func TestLintIgnoreList(t *testing.T) {
	diagnostics := GetDiagnostics(uri.File("../../testdata/example"), chartutil.Values{}, []string{"icon is recommended"})
	assert.NotEmpty(t, diagnostics)
	assert.Len(t, diagnostics, 1)
	assert.Empty(t, diagnostics[uri.File("../../testdata/example/Chart.yaml").Filename()])
}

func TestLintValuesJSONSchema(t *testing.T) {
	// Does currently not report schema validation errors
	diagnostics := GetDiagnostics(uri.File("../../testdata/example-json-schema"), chartutil.Values{}, []string{})
	assert.NotEmpty(t, diagnostics)
	assert.Len(t, diagnostics, 1)
	assert.Len(t, diagnostics[uri.File("../../testdata/example-json-schema/Chart.yaml").Filename()], 1)
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
	}, util.HelmLintConfig{Enabled: true})
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
	// Ensure a diagnostic notification is sent even if there are no diagnostics to remove old diagnostics after they are fixed
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
	}, util.HelmLintConfig{Enabled: true})

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

func TestLintNotificationsDisabled(t *testing.T) {
	chart := charts.Chart{
		RootURI:     uri.File("../../testdata/example"),
		ValuesFiles: &charts.ValuesFiles{},
	}
	diagnostics := GetDiagnosticsNotifications(&chart, &document.TemplateDocument{
		Document: document.Document{
			URI: uri.File("../../testdata/example/templates/deployment-no-templates.yaml"),
		},
	}, util.HelmLintConfig{Enabled: false})
	assert.Empty(t, diagnostics)
}
