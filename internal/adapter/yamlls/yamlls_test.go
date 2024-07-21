package yamlls

import (
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

	connector.documents = &lsplocal.DocumentStore{}
	yamlFile := uri.File("../../../testdata/example/templates/deployment.yaml")
	nonYamlFile := uri.File("../../../testdata/example/templates/_helpers.tpl")
	connector.documents.Store(yamlFile, util.DefaultConfig)
	connector.documents.Store(nonYamlFile, util.DefaultConfig)

	assert.True(t, connector.isRelevantFile(yamlFile))
	assert.False(t, connector.isRelevantFile(nonYamlFile))
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
