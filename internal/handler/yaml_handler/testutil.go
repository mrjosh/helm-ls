package yamlhandler

import (
	"os"
	"testing"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func setupYamlHandlerTest(t *testing.T, filepath string, loadTemplates bool) (handler *YamlHandler, fileContent string) {
	t.Helper()
	fileURI := uri.File(filepath)
	documents := document.NewDocumentStore()

	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatal("Could not read test file", err)
	}
	d := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:        fileURI,
			LanguageID: "",
			Version:    0,
			Text:       string(content),
		},
	}
	documents.DidOpenYamlDocument(&d, util.DefaultConfig)
	addChartCallback := func(chart *charts.Chart) {
		if loadTemplates {
			documents.LoadDocsOnNewChart(chart, util.DefaultConfig)
		}
	}
	chartStore := charts.NewChartStore(uri.File("."), charts.NewChart, addChartCallback)
	h := &YamlHandler{
		chartStore:      chartStore,
		documents:       documents,
		yamllsConnector: &yamlls.Connector{},
	}
	return h, string(content)
}
