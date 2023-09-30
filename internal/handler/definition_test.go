package handler

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	gotemplate "github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	yamlv3 "gopkg.in/yaml.v3"
)

var testFileContent = `
{{ $variable := "text" }} # line 1
{{ $variable }}           # line 2

{{ $someOther := "text" }}# line 4
{{ $variable }}           # line 5

{{ range $index, $element := pipeline }}{{ $index }}{{ $element }}{{ end }} # line 7
{{ .Values.foo }} # line 8
{{ .Values.something.nested }} # line 9
`

var testDocumentTemplateURI = uri.URI("file:///test.yaml")
var testValuesURI = uri.URI("file:///values.yaml")
var valuesContent = `
foo: bar
something: 
  nested: false
`

func genericDefinitionTest(t *testing.T, position lsp.Position, expectedLocation lsp.Location, expectedError error) {
	var node yamlv3.Node
	var err = yamlv3.Unmarshal([]byte(valuesContent), &node)
	if err != nil {
		t.Fatal(err)
	}
	handler := &langHandler{
		linterName: "helm-lint",
		connPool:   nil,
		documents:  nil,
		values:     make(map[string]interface{}),
		valueNode:  node,
		projectFiles: ProjectFiles{
			ValuesFile: "/values.yaml",
			ChartFile:  "",
		},
	}

	parser := sitter.NewParser()
	parser.SetLanguage(gotemplate.GetLanguage())
	tree, _ := parser.ParseCtx(context.Background(), nil, []byte(testFileContent))
	doc := &lsplocal.Document{
		Content: testFileContent,
		URI:     testDocumentTemplateURI,
		Ast:     tree,
	}

	location, err := handler.definitionAstParsing(doc, position)

	if err != nil && err.Error() != expectedError.Error() {
		t.Errorf("expected %v, got %v", expectedError, err)
	}

	if reflect.DeepEqual(location, expectedLocation) == false {
		t.Errorf("expected %v, got %v", expectedLocation, location)
	}
}

// Input:
// {{ $variable }}           # line 2
// -----|                    # this line incides the coursor position for the test
func TestDefinitionVariable(t *testing.T) {
	genericDefinitionTest(t, lsp.Position{Line: 2, Character: 8}, lsp.Location{
		URI: testDocumentTemplateURI,
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      1,
				Character: 3,
			},
		},
	}, nil)
}

func TestDefinitionNotImplemented(t *testing.T) {
	genericDefinitionTest(t, lsp.Position{Line: 1, Character: 1}, lsp.Location{
		Range: lsp.Range{},
	},
		fmt.Errorf("Definition not implemented for node type %s", "{{"))
}

// Input:
// {{ range $index, $element := pipeline }}{{ $index }}{{ $element }}{{ end }} # line 7
// -----------------------------------------------------------|
// Expected:
// {{ range $index, $element := pipeline }}{{ $index }}{{ $element }}{{ end }} # line 7
// -----------------|
func TestDefinitionRange(t *testing.T) {
	genericDefinitionTest(t, lsp.Position{Line: 7, Character: 60}, lsp.Location{
		URI: testDocumentTemplateURI,
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      7,
				Character: 17,
			},
		},
	}, nil)
}

// Input:
// {{ .Values.foo }} # line 8
// ------------|
func TestDefinitionValue(t *testing.T) {
	genericDefinitionTest(t, lsp.Position{Line: 8, Character: 13}, lsp.Location{
		URI: testValuesURI,
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      1,
				Character: 0,
			},
		},
	}, nil)
}

// Input:
// {{ .Values.something.nested }} # line 9
// ----------------------|
func TestDefinitionValueNested(t *testing.T) {
	genericDefinitionTest(t, lsp.Position{Line: 9, Character: 26}, lsp.Location{
		URI: testValuesURI,
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      3,
				Character: 2,
			},
		},
	}, nil)
}

// {{ .Values.foo }} # line 8
// ------|
func TestDefinitionValueFile(t *testing.T) {
	genericDefinitionTest(t, lsp.Position{Line: 8, Character: 7}, lsp.Location{
		URI: testValuesURI,
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      0,
				Character: 0,
			},
		},
	}, nil)
}
