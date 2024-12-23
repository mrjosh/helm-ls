package templatehandler

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	yamlv3 "gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
)

var (
	testFileContent = `
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

	documents := document.NewDocumentStore()
	fileURI := testDocumentTemplateURI
	rootUri := uri.File("/")

	testChart := &charts.Chart{
		ChartMetadata: &charts.ChartMetadata{},
		ValuesFiles: &charts.ValuesFiles{
			MainValuesFile: &charts.ValuesFile{
				Values:    make(map[string]interface{}),
				ValueNode: node,
				URI:       testValuesURI,
			},
			AdditionalValuesFiles: []*charts.ValuesFile{},
		},
		RootURI:   "",
		HelmChart: &chart.Chart{},
	}
	d := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:        fileURI,
			LanguageID: "",
			Version:    0,
			Text:       string(testFileContent),
		},
	}
	documents.DidOpenTemplateDocument(&d, util.DefaultConfig)
	chartStore := charts.NewChartStore(rootUri, charts.NewChart, addChartCallback)
	chartStore.Charts = map[uri.URI]*charts.Chart{rootUri: testChart}
	h := &TemplateHandler{
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
	documents := document.NewDocumentStore()
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
	documents.DidOpenTemplateDocument(&d, util.DefaultConfig)
	chartStore := charts.NewChartStore(rootUri, charts.NewChart, addChartCallback)
	chartStore.Charts = map[uri.URI]*charts.Chart{rootUri: chart}
	h := &TemplateHandler{
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

func TestDefinitionSingleLine(t *testing.T) {
	testCases := []struct {
		// defines a definition test where ^ is the position where the defintion is triggered
		// and §result§ marks the range of the result
		templateWithMarks string
	}{
		{"{{ §$test := 1§ }} {{ $te^st }}"},
		{"{{ §$test := .Values.test§ }} {{ $te^st.with.selectorexpression }}"},
		{"{{ §$test := $.Values.test§ }} {{ $te^st.with.selectorexpression. }}"},
		{"{{ §$test := .Values.test§ }} {{ $te^st }}"},
		{"{{ range §$test := .Values.test§ }} {{ $te^st }} {{ end }}"},
		{"{{ range §$test := $.Values.test§ }} {{ $te^st.something }} {{ end }}"},
		{"{{ range §$test := $.Values.test§ }} {{ $te^st. }} {{ end }}"},
		{"{{ range §$test := $.Values.test§ }} {{ if not $te^st }} y {{ else }} n {{ end }}"},
	}
	for _, tc := range testCases {
		t.Run(tc.templateWithMarks, func(t *testing.T) {
			col := strings.Index(tc.templateWithMarks, "^")
			buf := strings.Replace(tc.templateWithMarks, "^", "", 1)
			pos := protocol.Position{Line: 0, Character: uint32(col - 3)}
			expectedColStart := strings.Index(buf, "§")
			buf = strings.Replace(buf, "§", "", 1)
			expectedColEnd := strings.Index(buf, "§")
			buf = strings.Replace(buf, "§", "", 1)

			documents := document.NewDocumentStore()
			fileURI := testDocumentTemplateURI
			rootUri := uri.File("/")

			d := lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI:  fileURI,
					Text: buf,
				},
			}
			documents.DidOpenTemplateDocument(&d, util.DefaultConfig)
			h := &TemplateHandler{
				chartStore: charts.NewChartStore(rootUri, charts.NewChart, addChartCallback),
				documents:  documents,
			}

			locations, err := h.Definition(context.TODO(), &lsp.DefinitionParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{URI: fileURI},
					Position:     pos,
				},
			})

			assert.NoError(t, err)

			assert.Contains(t, locations, lsp.Location{
				URI: testDocumentTemplateURI,
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      0,
						Character: uint32(expectedColStart),
					},
					End: lsp.Position{
						Line:      0,
						Character: uint32(expectedColEnd),
					},
				},
			})
		})
	}
}
