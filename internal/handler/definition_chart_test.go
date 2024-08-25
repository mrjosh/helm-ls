package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

var (
	rootUri           = uri.File("../../testdata/dependenciesExample/")
	fileURI           = uri.File("../../testdata/dependenciesExample/templates/deployment.yaml")
	fileURIInSubchart = uri.File("../../testdata/dependenciesExample/charts/subchartexample/templates/subchart.yaml")
)

type testCase struct {
	// Must be content of a line in the file fileURI
	templateLineWithMarker string
	expectedFile           string
	expectedFileCount      int
	expectedStartPosition  lsp.Position
	expectedError          error
	inSubchart             bool
}

// Test definition on a real chart found in $rootUri
func TestDefinitionChart(t *testing.T) {
	testCases := []testCase{
		{
			`{{ include "common.na^mes.name" . }}`,
			"charts/.helm_ls_cache/common/templates/_names.tpl",
			1,
			lsp.Position{Line: 9, Character: 0},
			nil,
			false,
		},
		{
			`{{- include "dependeciesEx^ample.labels" . | nindent 4 }}`,
			"templates/_helpers.tpl",
			1,
			lsp.Position{Line: 35, Character: 0},
			nil,
			false,
		},
		{
			`{{ .Values.gl^obal.subchart }}`,
			"values.yaml",
			2,
			lsp.Position{Line: 7, Character: 0},
			nil,
			false,
		},
		{
			`{{ .Values.gl^obal.subchart }}`,
			"charts/subchartexample/values.yaml",
			2,
			lsp.Position{Line: 0, Character: 0},
			nil,
			false,
		},
		{
			`{{ .Values.common.exa^mpleValue }}`,
			"charts/.helm_ls_cache/common/values.yaml",
			1,
			// this tests, that the file also contains comments
			lsp.Position{Line: 7, Character: 0},
			nil,
			false,
		},
		{
			`{{ .Values.comm^on.exampleValue }}`,
			"charts/.helm_ls_cache/common/values.yaml",
			1,
			lsp.Position{Line: 7, Character: 0},
			nil,
			false,
		},
		{
			`{{ .Values.subch^artexample.subchartWithoutGlobal }}`,
			"values.yaml",
			2,
			lsp.Position{Line: 49, Character: 0},
			nil,
			false,
		},
		{
			`{{ .Values.subch^artexample.subchartWithoutGlobal }}`,
			"charts/subchartexample/values.yaml",
			2,
			lsp.Position{Line: 0, Character: 0},
			nil,
			false,
		},
		{
			`{{ .Values.subchartexample.subchartWith^outGlobal }}`,
			"values.yaml",
			2,
			lsp.Position{Line: 50, Character: 2},
			nil,
			false,
		},
		{
			`{{ .Values.subchartexample.subchart^WithoutGlobal }}`,
			"charts/subchartexample/values.yaml",
			2,
			lsp.Position{Line: 2, Character: 0},
			nil,
			false,
		},
		{
			`{{ .Cha^rt.Name }}`,
			"Chart.yaml",
			1,
			lsp.Position{Line: 0, Character: 0},
			nil,
			false,
		},
		{
			`{{ .Chart.Na^me }}`,
			"Chart.yaml",
			1,
			lsp.Position{Line: 1, Character: 0},
			nil,
			false,
		},
		{
			`{{ .Values.subchart^WithoutGlobal }}`,
			"charts/subchartexample/values.yaml",
			2,
			lsp.Position{Line: 2, Character: 0},
			nil,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run("Definition on "+tc.templateLineWithMarker, func(t *testing.T) {
			uri := fileURI
			if tc.inSubchart {
				uri = fileURIInSubchart
			}
			fileContent, err := os.ReadFile(uri.Filename())
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(string(fileContent), "\n")

			pos, found := getPosition(tc, lines)
			if !found {
				t.Fatal(fmt.Sprintf("%s is not in the file %s", tc.templateLineWithMarker, fileURI.Filename()))
			}

			documents := lsplocal.NewDocumentStore()

			chart := charts.NewChart(rootUri, util.DefaultConfig.ValuesFilesConfig)

			addChartCallback := func(chart *charts.Chart) {}
			chartStore := charts.NewChartStore(rootUri, charts.NewChart, addChartCallback)
			_, err = chartStore.GetChartForURI(rootUri)
			h := &langHandler{
				chartStore:      chartStore,
				documents:       documents,
				yamllsConnector: &yamlls.Connector{},
				helmlsConfig:    util.DefaultConfig,
			}

			assert.NoError(t, err)

			h.LoadDocsOnNewChart(chart)

			locations, err := h.Definition(context.TODO(), &lsp.DefinitionParams{
				TextDocumentPositionParams: lsp.TextDocumentPositionParams{
					TextDocument: lsp.TextDocumentIdentifier{URI: uri},
					Position:     pos,
				},
			})

			assert.Equal(t, tc.expectedError, err)
			assert.Len(t, locations, tc.expectedFileCount)

			// find the location with the correct file path
			foundLocation := false
			for _, location := range locations {
				if location.URI.Filename() == filepath.Join(rootUri.Filename(), tc.expectedFile) {
					locations = []lsp.Location{location}
					foundLocation = true
					break
				}
			}

			assert.True(t, foundLocation, fmt.Sprintf("Did not find a result with the expected file path %s ", filepath.Join(rootUri.Filename(), tc.expectedFile)))

			if len(locations) > 0 {
				assert.Equal(t, filepath.Join(rootUri.Filename(), tc.expectedFile), locations[0].URI.Filename())
				assert.Equal(t, tc.expectedStartPosition, locations[0].Range.Start)
			}

			for _, location := range locations {
				assert.FileExists(t, location.URI.Filename())
			}

			os.RemoveAll(filepath.Join(rootUri.Filename(), "charts", charts.DependencyCacheFolder))
		})
	}
}

func getPosition(tC testCase, lines []string) (lsp.Position, bool) {
	col := strings.Index(tC.templateLineWithMarker, "^")
	buf := strings.Replace(tC.templateLineWithMarker, "^", "", 1)
	line := uint32(0)
	found := false

	for i, v := range lines {
		if strings.Contains(v, buf) {
			found = true
			line = uint32(i)
			col = col + strings.Index(v, buf)
			break
		}
	}
	pos := lsp.Position{Line: line, Character: uint32(col)}
	return pos, found
}
