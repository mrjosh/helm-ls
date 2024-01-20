package charts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/pkg/chartutil"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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
