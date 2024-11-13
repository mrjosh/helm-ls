package yamlhandler

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/jsonrpc2"
)

func (h *YamlHandler) Configure(ctx context.Context, helmlsConfig util.HelmlsConfiguration) {
	h.configureYamlls(ctx, helmlsConfig)
}

func (h *YamlHandler) configureYamlls(ctx context.Context, helmlsConfig util.HelmlsConfiguration) {
	config := helmlsConfig

	customHandler := yamlls.CustomHandler{
		Handler: h.CustomHandler,
		PostInitialize: func(ctx context.Context, conn jsonrpc2.Conn) error {
			return conn.Notify(ctx, "yaml/registerCustomSchemaRequest", nil)
		},
	}

	connector := yamlls.NewConnector(ctx,
		config.YamllsConfiguration,
		h.client,
		h.documents,
		customHandler)
	h.setYamllsConnector(connector)

	err := h.yamllsConnector.CallInitialize(ctx, h.chartStore.RootURI)
	if err != nil {
		logger.Error("Error initializing yamlls", err)
	}

	h.yamllsConnector.InitiallySyncOpenYamlDocuments(h.documents.GetAllYamlDocs())
}
