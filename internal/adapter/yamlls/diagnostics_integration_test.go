package yamlls

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// must be relative to this file
var TEST_DATA_DIR = "../../../testdata/charts/bitnami/"

type jsonRpcDiagnostics struct {
	Params  lsp.PublishDiagnosticsParams `json:"params"`
	Jsonrpc string                       `json:"jsonrpc"`
	Method  string                       `json:"method"`
}

type readWriteCloseMock struct {
	diagnosticsChan chan lsp.PublishDiagnosticsParams
}

func (proc readWriteCloseMock) Read(p []byte) (int, error) {
	return 1, nil
}

func (proc readWriteCloseMock) Write(p []byte) (int, error) {
	if strings.HasPrefix(string(p), "Content-Length: ") {
		return 1, nil
	}
	var diagnostics jsonRpcDiagnostics
	json.NewDecoder(strings.NewReader(string(p))).Decode(&diagnostics)

	proc.diagnosticsChan <- diagnostics.Params
	return 1, nil
}

func (proc readWriteCloseMock) Close() error {
	return nil
}

func readTestFiles(dir string, channel chan<- lsp.DidOpenTextDocumentParams, doneChan chan<- int) {
	libRegEx, e := regexp.Compile(".*(/|\\\\)templates(/|\\\\).*\\.ya?ml")
	if e != nil {
		log.Fatal(e)
		return
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Fatal(err)
		return
	}

	count := 0
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if d.Type().IsRegular() && libRegEx.MatchString(path) {
			contentBytes, _ := os.ReadFile(path)
			count++
			channel <- lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI:        uri.File(path),
					LanguageID: "",
					Version:    0,
					Text:       string(contentBytes),
				},
			}
		}
		return nil
	})
	doneChan <- count
}

func sendTestFilesToYamlls(documents *lsplocal.DocumentStore, yamllsConnector *Connector,
	doneReadingFilesChan <-chan int,
	doneSendingFilesChan chan<- int,
	filesChan <-chan lsp.DidOpenTextDocumentParams,
) {
	ownCount := 0
	for {
		select {
		case d := <-filesChan:
			documents.DidOpen(d, util.DefaultConfig)
			tree := lsplocal.ParseAst(nil, d.TextDocument.Text)
			yamllsConnector.DocumentDidOpen(tree, d)
			ownCount++
		case count := <-doneReadingFilesChan:
			if count != ownCount {
				log.Fatal("Count mismatch: ", count, " != ", ownCount)
			}
			doneSendingFilesChan <- count
			return
		}
	}
}

func TestYamllsDiagnosticsIntegration(t *testing.T) {
	diagnosticsChan := make(chan lsp.PublishDiagnosticsParams)
	doneReadingFilesChan := make(chan int)
	doneSendingFilesChan := make(chan int)

	dir := t.TempDir()
	documents := lsplocal.NewDocumentStore()
	con := jsonrpc2.NewConn(jsonrpc2.NewStream(readWriteCloseMock{diagnosticsChan}))
	config := util.DefaultConfig.YamllsConfiguration

	yamllsSettings := util.DefaultYamllsSettings
	// disabling yamlls schema store improves performance and
	// removes all schema diagnostics that are not caused by the yaml trimming
	yamllsSettings.Schemas = make(map[string]string)
	yamllsSettings.YamllsSchemaStoreSettings = util.YamllsSchemaStoreSettings{
		Enable: false,
	}
	config.YamllsSettings = yamllsSettings
	yamllsConnector := NewConnector(config, con, documents)

	if yamllsConnector.Conn == nil {
		t.Fatal("Could not connect to yaml-language-server")
	}

	yamllsConnector.CallInitialize(uri.File(dir))

	didOpenChan := make(chan lsp.DidOpenTextDocumentParams)
	go readTestFiles(TEST_DATA_DIR, didOpenChan, doneReadingFilesChan)
	go sendTestFilesToYamlls(documents,
		yamllsConnector, doneReadingFilesChan, doneSendingFilesChan, didOpenChan)

	sentCount, diagnosticsCount := 0, 0
	receivedDiagnostics := make(map[uri.URI]lsp.PublishDiagnosticsParams)

	afterCh := time.After(600 * time.Second)
	for {
		if sentCount != 0 && len(receivedDiagnostics) == sentCount {
			fmt.Println("All files checked")
			break
		}
		select {
		case d := <-diagnosticsChan:
			receivedDiagnostics[d.URI] = d
			if len(d.Diagnostics) > 0 {
				diagnosticsCount++
				fmt.Printf("Got diagnostic in %s diagnostics: %v \n", d.URI.Filename(), d.Diagnostics)
			}
		case <-afterCh:
			t.Fatal("Timed out waiting for diagnostics")
		case count := <-doneSendingFilesChan:
			sentCount = count
		}
	}

	fmt.Printf("Checked %d files, found %d diagnostics\n", sentCount, diagnosticsCount)
	assert.LessOrEqual(t, diagnosticsCount, 23)
	assert.Equal(t, 2368, sentCount, "Count of files in test data not equal to actual count")
}
