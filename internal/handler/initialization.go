package handler

import (
	"context"
	"os"

	"github.com/gobwas/glob"
	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/sirupsen/logrus"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (h *langHandler) Initialize(ctx context.Context, params *lsp.InitializeParams) (result *lsp.InitializeResult, err error) {
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
	h.chartStore = charts.NewChartStore(workspaceURI, charts.NewChart, h.AddChartCallback)

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

func (h *langHandler) Initialized(ctx context.Context, _ *lsp.InitializedParams) (err error) {
	h.retrieveWorkspaceConfiguration(ctx)
	return nil
}

func (h *langHandler) initializationWithConfig(ctx context.Context) {
	configureLogLevel(h.helmlsConfig)
	h.chartStore.SetValuesFilesConfig(h.helmlsConfig.ValuesFilesConfig)
	h.configureYamlls(ctx)
}

func (h *langHandler) configureYamlsEnabledGlob() {
	globObject, err := glob.Compile(h.helmlsConfig.YamllsConfiguration.EnabledForFilesGlob)
	if err != nil {
		logger.Error("Error compiling glob for yamlls EnabledForFilesGlob", err)
		globObject = util.DefaultConfig.YamllsConfiguration.EnabledForFilesGlobObject
	}
	h.helmlsConfig.YamllsConfiguration.EnabledForFilesGlobObject = globObject
}

func (h *langHandler) configureYamlls(ctx context.Context) {
	config := h.helmlsConfig
	if config.YamllsConfiguration.Enabled {
		h.configureYamlsEnabledGlob()
		h.yamllsConnector = yamlls.NewConnector(ctx, config.YamllsConfiguration, h.client, h.documents)
		err := h.yamllsConnector.CallInitialize(ctx, h.chartStore.RootURI)
		if err != nil {
			logger.Error("Error initializing yamlls", err)
		}

		h.yamllsConnector.InitiallySyncOpenDocuments(h.documents.GetAllTemplateDocs())
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

func (h *langHandler) AddChartCallback(chart *charts.Chart) {
	h.NewChartWithWatchedFiles(chart)
	go h.LoadDocsOnNewChart(chart)
}
