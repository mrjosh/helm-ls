package charts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestNewValuesFiles(t *testing.T) {
	tempDir := t.TempDir()

	valuesContent := `foo: bar`
	additionalValuesContent := `bar: baz`

	_ = os.WriteFile(filepath.Join(tempDir, "values.yaml"), []byte(valuesContent), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "values-additional.yaml"), []byte(additionalValuesContent), 0o644)

	valuesFiles := charts.NewValuesFiles(uri.New("file://"+tempDir), "values.yaml", "", "values*.yaml")

	assert.Equal(t, "bar", valuesFiles.MainValuesFile.Values["foo"])
	assert.Equal(t, "baz", valuesFiles.AdditionalValuesFiles[0].Values["bar"])
	assert.Equal(t, 1, len(valuesFiles.AdditionalValuesFiles))
}

func TestGetPositionsForValue(t *testing.T) {
	tempDir := t.TempDir()

	valuesContent := `foo: bar`
	additionalValuesContent := `
other: value
foo: baz`

	_ = os.WriteFile(filepath.Join(tempDir, "values.yaml"), []byte(valuesContent), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "values-additional.yaml"), []byte(additionalValuesContent), 0o644)

	valuesFiles := charts.NewValuesFiles(uri.New("file://"+tempDir), "values.yaml", "", "values*.yaml")

	assert.Equal(t, []lsp.Location{
		{
			URI:   uri.New("file://" + filepath.Join(tempDir, "values.yaml")),
			Range: lsp.Range{Start: lsp.Position{Line: 0, Character: 0}, End: lsp.Position{Line: 0, Character: 0}},
		},
		{
			URI:   uri.New("file://" + filepath.Join(tempDir, "values-additional.yaml")),
			Range: lsp.Range{Start: lsp.Position{Line: 2, Character: 0}, End: lsp.Position{Line: 2, Character: 0}},
		},
	}, valuesFiles.GetPositionsForValue([]string{"foo"}))
}

func TestNewValuesFileForLintOverlay(t *testing.T) {
	tempDir := t.TempDir()

	valuesContent := `foo: bar`
	additionalValuesContent := `bar: baz`

	_ = os.WriteFile(filepath.Join(tempDir, "values.yaml"), []byte(valuesContent), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "values-additional.yaml"), []byte(additionalValuesContent), 0o644)

	valuesFiles := charts.NewValuesFiles(uri.New("file://"+tempDir), "values.yaml", "values-additional.yaml", "values*.yaml")

	assert.Equal(t, "baz", valuesFiles.OverlayValuesFile.Values["bar"])
}

func TestNewValuesFileForLintOverlayNewFile(t *testing.T) {
	tempDir := t.TempDir()

	valuesContent := `foo: bar`
	additionalValuesContent := `bar: baz`

	_ = os.WriteFile(filepath.Join(tempDir, "values.yaml"), []byte(valuesContent), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "values-additional.yaml"), []byte(additionalValuesContent), 0o644)

	valuesFiles := charts.NewValuesFiles(uri.New("file://"+tempDir), "values.yaml", "values-additional.yaml", "")

	assert.Equal(t, "baz", valuesFiles.OverlayValuesFile.Values["bar"])
}

func TestNewValuesFileForLintOverlayPicksFirst(t *testing.T) {
	tempDir := t.TempDir()

	valuesContent := `foo: bar`
	additionalValuesContent := `bar: baz`

	_ = os.WriteFile(filepath.Join(tempDir, "values.yaml"), []byte(valuesContent), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "values-additional.yaml"), []byte(additionalValuesContent), 0o644)

	valuesFiles := charts.NewValuesFiles(uri.New("file://"+tempDir), "values.yaml", "", "values*.yaml")

	assert.Equal(t, "baz", valuesFiles.OverlayValuesFile.Values["bar"])
}
