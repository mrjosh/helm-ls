package yamlhandler

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/util"
)

func (h *YamlHandler) Configure(ctx context.Context, helmlsConfig util.HelmlsConfiguration) {
	h.configureYamlls(ctx, helmlsConfig.YamllsConfiguration)
}

func (h *YamlHandler) configureYamlls(ctx context.Context, config util.YamllsConfiguration) {
	// NOTE: For now we disable yamlls diagnostics since we expect the user
	// to also run yaml-language-server separately for diagnostics
	config.DiagnosticsEnabled = false

	connector := yamlls.NewConnector(ctx,
		config,
		h.client,
		h.documents,
		yamlls.NewCustomSchemaHandler(
			yamlls.NewCustomSchemaProviderHandler(h.CustomSchemaProvider),
		),
	)

	h.setYamllsConnector(connector)

	err := h.yamllsConnector.CallInitialize(ctx, h.chartStore.RootURI)
	if err != nil {
		logger.Error("Error initializing yamlls", err)
	}

	h.yamllsConnector.InitiallySyncOpenYamlDocuments(h.documents.GetAllYamlDocs())
}
