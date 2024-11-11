//go:build integration

package yamlls

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// must be relative to this file
var TEST_DATA_DIR = "../../../testdata/charts/bitnami/"

func readTestFiles(dir string, channel chan<- string, doneChan chan<- int) {
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
			count++
			channel <- path
		}
		return nil
	})
	doneChan <- count
}

func sendTestFilesToYamlls(documents *document.DocumentStore, yamllsConnector *Connector,
	doneReadingFilesChan <-chan int,
	doneSendingFilesChan chan<- int,
	filesChan <-chan string,
) {
	ownCount := 0
	for {
		select {
		case d := <-filesChan:
			openFile(&testing.T{}, documents, d, yamllsConnector)
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
	doneReadingFilesChan := make(chan int)
	doneSendingFilesChan := make(chan int)

	config := util.DefaultConfig.YamllsConfiguration

	yamllsSettings := util.DefaultYamllsSettings
	// disabling yamlls schema store improves performance and
	// removes all schema diagnostics that are not caused by the yaml trimming
	yamllsSettings.Schemas = make(map[string]string)
	yamllsSettings.YamllsSchemaStoreSettings = util.YamllsSchemaStoreSettings{
		Enable: false,
	}
	config.YamllsSettings = yamllsSettings
	yamllsConnector, documents, diagnosticsChan := getYamlLsConnector(t, config)

	didOpenChan := make(chan string)
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

func TestYamllsDiagnosticsIntegrationWithSchema(t *testing.T) {
	t.Parallel()
	diagnosticsChan := make(chan lsp.PublishDiagnosticsParams)

	config := util.DefaultConfig.YamllsConfiguration
	yamllsConnector, documents, diagnosticsChan := getYamlLsConnector(t, config)
	file := filepath.Join("..", "..", "..", "testdata", "example", "templates", "service.yaml")
	openFile(t, documents, file, yamllsConnector)

	expected := lsp.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      1.0,
				Character: 0,
			},
			End: protocol.Position{
				Line:      1,
				Character: 5,
			},
		},
		Severity:           1,
		Code:               0.0,
		CodeDescription:    nil,
		Source:             "yaml-schema: https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/v1.22.4-standalone-strict/_definitions.json",
		Message:            "Yamlls: Property wrong is not allowed.",
		Tags:               nil,
		RelatedInformation: nil,
		Data: map[string]interface{}{
			"properties": []interface{}{
				"status",
			},
			"schemaUri": []interface{}{
				"https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/v1.22.4-standalone-strict/_definitions.json",
			},
		},
	}

	diagnostic := []lsp.Diagnostic{}
	afterCh := time.After(10 * time.Second)
	for {
		if len(diagnostic) > 0 {
			break
		}
		select {
		case d := <-diagnosticsChan:
			diagnostic = append(diagnostic, d.Diagnostics...)
		case <-afterCh:
			t.Fatal("Timed out waiting for diagnostics")
		}
	}

	assert.Contains(t, diagnostic, expected)
	assert.Len(t, diagnostic, 1)
}
