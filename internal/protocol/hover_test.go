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

	expected := fmt.Sprintf(`### file3.yaml
%s
value3
%s
### file2.yaml
%s
value2
%s
### file1.yaml
%s
value1
%s
`, "```yaml", "```", "```yaml", "```", "```yaml", "```")

	formatted := results.FormatYaml(rootURI)
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

	formatted := results.FormatYaml(rootURI)
	assert.Equal(t, expected, formatted, "The formatted output should match the expected output")
}

func TestHoverResultsWithFiles_Format_DifferenPath(t *testing.T) {
	rootURI := uri.New("file:///home/user/project")

	results := HoverResultsWithFiles{
		{Value: "value", URI: uri.New("file:///invalid/uri")},
	}

	expected := fmt.Sprintf(`### %s
%s
value
%s
`, filepath.Join("..", "..", "..", "invalid", "uri"), "```yaml", "```")
	formatted := results.FormatYaml(rootURI)
	assert.Equal(t, expected, formatted, "The formatted output should match the expected output")
}

func TestHoverResultWithFile_WithHelmCode(t *testing.T) {
	hoverResult := HoverResultsWithFiles{
		{
			Value: "some helm code",
			URI:   uri.New("file:///home/user/project/file1.yaml"),
		},
	}

	expectedValue := "### file1.yaml\n```helm\nsome helm code\n```\n"
	assert.Equal(t, expectedValue, hoverResult.FormatHelm(uri.New("file:///home/user/project")), "The value should be formatted with Helm code block")
}
