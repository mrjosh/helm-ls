package handler

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/mrjosh/helm-ls/pkg/chartutil"
	"github.com/sirupsen/logrus"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (h *langHandler) handleInitialize(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {

	var params lsp.InitializeParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	if len(params.WorkspaceFolders) == 0 {
		return errors.New("length WorkspaceFolders is 0")
	}

	workspaceURI, err := uri.Parse(params.WorkspaceFolders[0].URI)
	h.yamllsConnector.CallInitialize(workspaceURI)

	h.projectFiles = NewProjectFiles(workspaceURI, "")

	vals, err := chartutil.ReadValuesFile(h.projectFiles.ValuesFile)
	if err != nil {
		logger.Println("Error loading values.yaml file", err)
	}
	h.values = vals

	chartMetadata, err := chartutil.LoadChartfile(h.projectFiles.ChartFile)
	if err != nil {
		logger.Println("Error loading Chart.yaml file", err)
	}
	h.chartMetadata = *chartMetadata
	valueNodes, err := chartutil.ReadYamlFileToNode(h.projectFiles.ValuesFile)
	if err != nil {
		logger.Println("Error loading values.yaml file", err)
	}
	h.valueNode = valueNodes

	chartNode, err := chartutil.ReadYamlFileToNode(h.projectFiles.ChartFile)
	if err != nil {
		logger.Println("Error loading Chart.yaml file", err)
	}
	h.chartNode = chartNode

	return reply(ctx, lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: lsp.TextDocumentSyncOptions{
				Change:    lsp.TextDocumentSyncKindIncremental,
				OpenClose: true,
				Save: &lsp.SaveOptions{
					IncludeText: true,
				},
			},
			CompletionProvider: &lsp.CompletionOptions{
				TriggerCharacters: []string{".", "$."},
				ResolveProvider:   false,
			},
			HoverProvider:      true,
			DefinitionProvider: true,
		},
	}, nil)
}

func (h *langHandler) initializationWithConfig(ctx context.Context) {
	configureLogLevel(h.helmlsConfig)
	configureYamlls(h)
}

func configureYamlls(h *langHandler) {
	if h.helmlsConfig.YamllsConfiguration.Enabled {
		h.yamllsConnector = yamlls.NewConnector(h.helmlsConfig.YamllsConfiguration, h.connPool, h.documents)
		h.yamllsConnector.CallInitialize(h.projectFiles.RootURI)
		h.yamllsConnector.InitiallySyncOpenDocuments()
	}
}

func configureLogLevel(helmlsConfig util.HelmlsConfiguration) {
	if level, err := logrus.ParseLevel(helmlsConfig.LogLevel); err == nil {
		logger.SetLevel(level)
	} else {
		logger.Println("Error parsing log level", err)
	}
	if os.Getenv("LOG_LEVEL") == "debug" {
		logger.SetLevel(logrus.DebugLevel)
	}
}
