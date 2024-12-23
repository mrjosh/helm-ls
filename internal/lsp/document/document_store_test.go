package document

import (
	"testing"

	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
)

func TestIsYamlDocument(t *testing.T) {
	assert := assert.New(t)
	assert.True(IsYamllsEnabled(uri.File("test.yaml"), util.DefaultConfig.YamllsConfiguration))
	assert.False(IsYamllsEnabled(uri.File("test.tpl"), util.DefaultConfig.YamllsConfiguration))
	assert.True(IsYamllsEnabled(uri.File("../../testdata/example/templates/hpa.yaml"), util.DefaultConfig.YamllsConfiguration))
	assert.False(IsYamllsEnabled(uri.File("../../testdata/example/templates/_helpers.tpl"), util.DefaultConfig.YamllsConfiguration))
}
