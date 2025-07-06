package testutil

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

type ExpectedDefinitionResult struct {
	Filepath string
	// the expected line with the range marked using §§
	MarkedLine string
}

func (expected ExpectedDefinitionResult) GetLocation() (protocol.Location, error) {
	fileContent, err := os.ReadFile(expected.Filepath)
	if err != nil {
		return protocol.Location{}, err
	}
	foundRange, found := GetRangeOfMarkedLineInFile(string(fileContent), expected.MarkedLine, "§")

	if !found {
		return protocol.Location{}, fmt.Errorf("could not find marked line in file")
	}
	return protocol.Location{URI: uri.File(expected.Filepath), Range: foundRange}, nil
}

func AssertDefinitionResult(t *testing.T, actual []protocol.Location, expected []ExpectedDefinitionResult) {
	t.Helper()
	expectedLocations := []protocol.Location{}

	for _, e := range expected {
		location, err := e.GetLocation()
		if err != nil {
			t.Fatal(err)
		}
		expectedLocations = append(expectedLocations, location)
	}

	assert.ElementsMatch(t, expectedLocations, actual)
}
