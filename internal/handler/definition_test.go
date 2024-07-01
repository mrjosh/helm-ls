package handler

import (
	"context"
	"reflect"
	"testing"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
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

{{ range .Values.list }}
{{ . }} # line 12
{{ end }}
{{ range $index, $element := pipeline }}{{ $index }}{{ $element }}{{ end }} # line 14
`

var (
	testDocumentTemplateURI = uri.URI("file:///templates/test.yaml")
	testValuesURI           = uri.URI("file:///values.yaml")
	testOtherValuesURI      = uri.URI("file:///values.other.yaml")
	valuesContent           = `
foo: bar
something: 
  nested: false
list:
  - test
`
)

func genericDefinitionTest(t *testing.T, position lsp.Position, expectedLocations []lsp.Location, expectedError error) {
	var node yamlv3.Node
	err := yamlv3.Unmarshal([]byte(valuesContent), &node)
	if err != nil {
		t.Fatal(err)
	}

	documents := lsplocal.NewDocumentStore()
	fileURI := testDocumentTemplateURI
	rootUri := uri.File("/")

	chart := &charts.Chart{
		ChartMetadata: &charts.ChartMetadata{},
		ValuesFiles: &charts.ValuesFiles{
			MainValuesFile: &charts.ValuesFile{
				Values:    make(map[string]interface{}),
				ValueNode: node,
				URI:       testValuesURI,
			},
			AdditionalValuesFiles: []*charts.ValuesFile{},
		},
		RootURI: "",
	}
	d := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:        fileURI,
			LanguageID: "",
			Version:    0,
			Text:       string(testFileContent),
		},
	}
	documents.DidOpen(&d, util.DefaultConfig)
	chartStore := charts.NewChartStore(rootUri, charts.NewChart)
	chartStore.Charts = map[uri.URI]*charts.Chart{rootUri: chart}
	h := &langHandler{
		chartStore:      chartStore,
		documents:       documents,
		yamllsConnector: &yamlls.Connector{},
	}

	location, err := h.Definition(context.TODO(), &lsp.DefinitionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: fileURI},
			Position:     position,
		},
	})

	assert.Equal(t, expectedError, err)
	assert.Equal(t, expectedLocations, location)
}

// Input:
// {{ $variable }}           # line 2
// -----|                    # this line indicates the cursor position for the test
func TestDefinitionVariable(t *testing.T) {
	genericDefinitionTest(t, lsp.Position{Line: 2, Character: 8}, []lsp.Location{
		{
			URI: testDocumentTemplateURI,
			Range: lsp.Range{
				Start: lsp.Position{Line: 1, Character: 3},
				End:   lsp.Position{Line: 1, Character: 22},
			},
		},
	}, nil)
}

func TestDefinitionNotImplemented(t *testing.T) {
	genericDefinitionTest(t, lsp.Position{Line: 1, Character: 1}, nil, nil)
}

// Input:
// {{ range $index, $element := pipeline }}{{ $index }}{{ $element }}{{ end }} # line 7
// -----------------------------------------------------------|
// Expected:
// {{ range $index, $element := pipeline }}{{ $index }}{{ $element }}{{ end }} # line 7
// -----------------|
func TestDefinitionRange(t *testing.T) {
	genericDefinitionTest(t, lsp.Position{Line: 7, Character: 60}, []lsp.Location{
		{
			URI: testDocumentTemplateURI,
			Range: lsp.Range{
				Start: lsp.Position{Line: 7, Character: 17},
				End:   lsp.Position{Line: 7, Character: 37},
			},
		},
	}, nil)
}

// Input:
// {{ range $index, $element := pipeline }}{{ $index }}{{ $element }}{{ end }} # line 14
// ---------------------|
// Expected:
// {{ range $index, $element := pipeline }}{{ $index }}{{ $element }}{{ end }} # line 14
// -----------------|
func TestDefinitionRangeOnRedeclaration(t *testing.T) {
	genericDefinitionTest(t, lsp.Position{Line: 14, Character: 23}, []lsp.Location{
		{
			URI: testDocumentTemplateURI,
			Range: lsp.Range{
				Start: lsp.Position{Line: 14, Character: 17},
				End:   lsp.Position{Line: 14, Character: 37},
			},
		},
	}, nil)
}

// Input:
// {{ .Values.foo }} # line 8
// ------------|
func TestDefinitionValue(t *testing.T) {
	genericDefinitionTest(t, lsp.Position{Line: 8, Character: 13}, []lsp.Location{
		{
			URI: testValuesURI,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      1,
					Character: 0,
				},
				End: lsp.Position{
					Line:      1,
					Character: 0,
				},
			},
		},
	}, nil)
}

// Input:
// {{ range .Values.list }}
// {{ . }} # line 12
// ---|
func TestDefinitionValueInList(t *testing.T) {
	genericDefinitionTest(t, lsp.Position{Line: 12, Character: 3}, []lsp.Location{
		{
			URI: testValuesURI,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      4,
					Character: 0,
				},
				End: lsp.Position{
					Line:      4,
					Character: 0,
				},
			},
		},
	}, nil)
}

// Input:
// {{ . }} # line 9
// ----------------------|
func TestDefinitionValueNested(t *testing.T) {
	genericDefinitionTest(t, lsp.Position{Line: 9, Character: 26}, []lsp.Location{
		{
			URI: testValuesURI,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      3,
					Character: 2,
				},
				End: lsp.Position{
					Line:      3,
					Character: 2,
				},
			},
		},
	}, nil)
}

// {{ .Values.foo }} # line 8
// ------|
func TestDefinitionValueFile(t *testing.T) {
	genericDefinitionTest(t, lsp.Position{Line: 8, Character: 7}, []lsp.Location{
		{
			URI: testValuesURI,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      0,
					Character: 0,
				},
				End: lsp.Position{
					Line:      0,
					Character: 0,
				},
			},
		},
	}, nil)
}

func genericDefinitionTestMultipleValuesFiles(t *testing.T, position lsp.Position, expectedLocations []lsp.Location, expectedError error) {
	var node yamlv3.Node
	err := yamlv3.Unmarshal([]byte(valuesContent), &node)
	if err != nil {
		t.Fatal(err)
	}
	documents := lsplocal.NewDocumentStore()
	fileURI := testDocumentTemplateURI
	rootUri := uri.File("/")

	chart := &charts.Chart{
		ChartMetadata: &charts.ChartMetadata{},
		ValuesFiles: &charts.ValuesFiles{
			MainValuesFile: &charts.ValuesFile{
				Values:    make(map[string]interface{}),
				ValueNode: node,
				URI:       testValuesURI,
			},
			AdditionalValuesFiles: []*charts.ValuesFile{
				{
					Values:    make(map[string]interface{}),
					ValueNode: node,
					URI:       testOtherValuesURI,
				},
			},
		},
		RootURI: "",
	}
	d := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:        fileURI,
			LanguageID: "",
			Version:    0,
			Text:       string(testFileContent),
		},
	}
	documents.DidOpen(&d, util.DefaultConfig)
	chartStore := charts.NewChartStore(rootUri, charts.NewChart)
	chartStore.Charts = map[uri.URI]*charts.Chart{rootUri: chart}
	h := &langHandler{
		chartStore:      chartStore,
		documents:       documents,
		yamllsConnector: &yamlls.Connector{},
	}

	location, err := h.Definition(context.TODO(), &lsp.DefinitionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: fileURI},
			Position:     position,
		},
	})

	if err != nil && err.Error() != expectedError.Error() {
		t.Errorf("expected %v, got %v", expectedError, err)
	}

	if reflect.DeepEqual(location, expectedLocations) == false {
		t.Errorf("expected %v, got %v", expectedLocations, location)
	}
}

// {{ .Values.foo }} # line 8
// ------|
func TestDefinitionValueFileMulitpleValues(t *testing.T) {
	genericDefinitionTestMultipleValuesFiles(t, lsp.Position{Line: 8, Character: 7}, []lsp.Location{
		{
			URI: testValuesURI,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      0,
					Character: 0,
				},
				End: lsp.Position{
					Line:      0,
					Character: 0,
				},
			},
		}, {
			URI: testOtherValuesURI,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      0,
					Character: 0,
				},
				End: lsp.Position{
					Line:      0,
					Character: 0,
				},
			},
		},
	}, nil)
}
