package charts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chartutil"
)

func TestNewValuesFile(t *testing.T) {
	tempDir := t.TempDir()

	valuesContent := `foo: bar`
	_ = os.WriteFile(filepath.Join(tempDir, "values.yaml"), []byte(valuesContent), 0o644)
	valuesFile := charts.NewValuesFile(filepath.Join(tempDir, "values.yaml"))

	assert.Equal(t, "bar", valuesFile.Values["foo"])
	assert.NotEqual(t, yaml.Node{}, valuesFile.ValueNode)
}

func TestNewValuesFileFileNotFound(t *testing.T) {
	tempDir := t.TempDir()

	valuesFile := charts.NewValuesFile(filepath.Join(tempDir, "values.yaml"))

	assert.Equal(t, chartutil.Values{}, valuesFile.Values)
	assert.Equal(t, yaml.Node{}, valuesFile.ValueNode)
}

func TestReload(t *testing.T) {
	tempDir := t.TempDir()
	valuesContent := `foo: bar`
	_ = os.WriteFile(filepath.Join(tempDir, "values.yaml"), []byte(valuesContent), 0o644)
	valuesFile := charts.NewValuesFile(filepath.Join(tempDir, "values.yaml"))

	assert.Equal(t, "bar", valuesFile.Values["foo"])
	assert.NotEqual(t, yaml.Node{}, valuesFile.ValueNode)

	_ = os.WriteFile(filepath.Join(tempDir, "values.yaml"), []byte("foo: baz"), 0o644)
	valuesFile.Reload()
	assert.Equal(t, "baz", valuesFile.Values["foo"])
	assert.NotEqual(t, yaml.Node{}, valuesFile.ValueNode)
}

func TestGetContent(t *testing.T) {
	tempDir := t.TempDir()
	valuesContent := "foo: bar"
	_ = os.WriteFile(filepath.Join(tempDir, "values.yaml"), []byte(valuesContent), 0o644)
	valuesFile := charts.NewValuesFile(filepath.Join(tempDir, "values.yaml"))
	assert.Equal(t, valuesContent+"\n", valuesFile.GetContent())
}
