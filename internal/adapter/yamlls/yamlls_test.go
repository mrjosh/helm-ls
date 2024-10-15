package yamlls

import (
	"os"
	"testing"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
)

func TestIsRelevantFile(t *testing.T) {
	connector := Connector{
		config: util.YamllsConfiguration{
			Enabled: true,
		},
	}

	connector.documents = lsplocal.NewDocumentStore()
	yamlFile := "../../../testdata/example/templates/deployment.yaml"
	nonYamlFile := "../../../testdata/example/templates/_helpers.tpl"

	yamlFileContent, err := os.ReadFile(yamlFile)
	assert.NoError(t, err)

	nonYamlFileContent, err := os.ReadFile(nonYamlFile)
	assert.NoError(t, err)

	connector.documents.StoreTemplateDocument(yamlFile, yamlFileContent, util.DefaultConfig)
	connector.documents.StoreTemplateDocument(nonYamlFile, nonYamlFileContent, util.DefaultConfig)

	assert.True(t, connector.isRelevantFile(uri.File(yamlFile)))
	assert.False(t, connector.isRelevantFile(uri.File(nonYamlFile)))
}

func TestShouldRun(t *testing.T) {
	connector := Connector{
		config: util.YamllsConfiguration{
			Enabled: true,
		},
	}
	assert.False(t, connector.shouldRun(uri.File("test.yaml")))
	assert.False(t, connector.shouldRun(uri.File("_helpers.tpl")))
}
