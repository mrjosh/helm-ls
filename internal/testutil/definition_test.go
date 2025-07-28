package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestExpectedDefinitionResultGetLocation(t *testing.T) {
	sut := ExpectedLocationsResult{
		Filepath:   "../../testdata/example/values.yaml",
		MarkedLine: "§replicaCount§: 1",
	}

	actual, err := sut.GetLocation()

	assert.NoError(t, err)

	expected := protocol.Location{
		URI:   uri.File(sut.Filepath),
		Range: protocol.Range{Start: protocol.Position{Line: 4, Character: 0}, End: protocol.Position{Line: 4, Character: 12}},
	}

	assert.Equal(t, expected, actual)
}
