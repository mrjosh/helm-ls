package protocol

import (
	"fmt"
	"path/filepath"
	"testing"

	"go.lsp.dev/uri"

	"github.com/stretchr/testify/assert"
)

func TestHoverResultsWithFiles_Format(t *testing.T) {
	rootURI := uri.New("file:///home/user/project")

	results := HoverResultsWithFiles{
		{Value: "value1", URI: uri.New("file:///home/user/project/file1.yaml")},
		{Value: "value2", URI: uri.New("file:///home/user/project/file2.yaml")},
		{Value: "value3", URI: uri.New("file:///home/user/project/file3.yaml")},
	}

	expected := `### file3.yaml
value3

### file2.yaml
value2

### file1.yaml
value1

`

	formatted := results.Format(rootURI)
	assert.Equal(t, expected, formatted, "The formatted output should match the expected output")
}

func TestHoverResultsWithFiles_Format_EmptyValue(t *testing.T) {
	rootURI := uri.New("file:///home/user/project")

	results := HoverResultsWithFiles{
		{Value: "", URI: uri.New("file:///home/user/project/file1.yaml")},
	}
	expected := `### file1.yaml
""

`

	formatted := results.Format(rootURI)
	assert.Equal(t, expected, formatted, "The formatted output should match the expected output")
}

func TestHoverResultsWithFiles_Format_DifferenPath(t *testing.T) {
	rootURI := uri.New("file:///home/user/project")

	results := HoverResultsWithFiles{
		{Value: "value", URI: uri.New("file:///invalid/uri")},
	}

	expected := fmt.Sprintf(`### %s
value

`, filepath.Join("..", "..", "..", "invalid", "uri"))
	formatted := results.Format(rootURI)
	assert.Equal(t, expected, formatted, "The formatted output should match the expected output")
}

func TestHoverResultWithFile_WithHelmCode(t *testing.T) {
	hoverResult := HoverResultWithFile{
		Value: "some helm code",
		URI:   uri.New("file:///home/user/project/file1.yaml"),
	}.AsHelmCode()

	expectedValue := "```helm\nsome helm code\n```"
	assert.Equal(t, expectedValue, hoverResult.Value, "The value should be formatted with Helm code block")
}
