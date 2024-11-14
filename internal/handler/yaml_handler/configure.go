package yamlhandler

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/util"
)

func (h *YamlHandler) Configure(ctx context.Context, helmlsConfig util.HelmlsConfiguration) {
	h.configureYamlls(ctx, helmlsConfig)
}

func (h *YamlHandler) configureYamlls(ctx context.Context, helmlsConfig util.HelmlsConfiguration) {
	config := helmlsConfig

	connector := yamlls.NewConnector(ctx,
		config.YamllsConfiguration,
		h.client,
		h.documents,
		*yamlls.NewCustomSchemaHandler(
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
