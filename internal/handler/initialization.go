package handler

import (
	"context"
	"os"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/sirupsen/logrus"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (h *ServerHandler) Initialize(ctx context.Context, params *lsp.InitializeParams) (result *lsp.InitializeResult, err error) {
	var workspaceURI uri.URI

	if len(params.WorkspaceFolders) != 0 {
		workspaceURI, err = uri.Parse(params.WorkspaceFolders[0].URI)
		if err != nil {
			logger.Error("Error parsing workspace URI", err)
			return nil, err
		}
	} else {
		logger.Error("length WorkspaceFolders is 0, falling back to current working directory")
		workspaceURI = uri.File(".")
	}

	logger.Debug("Initializing chartStore")
	h.chartStore.SetRootURI(workspaceURI)

	logger.Debug("Initializing done")
	return &lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: lsp.TextDocumentSyncOptions{
				Change:    lsp.TextDocumentSyncKindIncremental,
				OpenClose: true,
				// ensure we get a save notification to update diagnostics
				Save: &lsp.SaveOptions{},
			},
			CompletionProvider: &lsp.CompletionOptions{
				TriggerCharacters: []string{".", "$."},
				ResolveProvider:   false,
			},
			HoverProvider:          true,
			DefinitionProvider:     true,
			ReferencesProvider:     true,
			DocumentSymbolProvider: true,
		},
	}, nil
}

func (h *ServerHandler) Initialized(ctx context.Context, _ *lsp.InitializedParams) (err error) {
	h.retrieveWorkspaceConfiguration(ctx)
	return nil
}

func (h *ServerHandler) initializationWithConfig(ctx context.Context) {
	configureLogLevel(h.helmlsConfig)
	h.chartStore.SetValuesFilesConfig(h.helmlsConfig.ValuesFilesConfig)

	for _, handler := range h.langHandlers {
		handler.Configure(ctx, h.helmlsConfig)
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

func (h *ServerHandler) AddChartCallback(chart *charts.Chart) {
	h.NewChartWithWatchedFiles(chart)
	go h.LoadDocsOnNewChart(chart)
}
