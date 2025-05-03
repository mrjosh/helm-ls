package templatehandler

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/util"
)

func (h *TemplateHandler) Configure(ctx context.Context, helmlsConfig util.HelmlsConfiguration) {
	h.configureYamlls(ctx, helmlsConfig.YamllsConfiguration)
}

func (h *TemplateHandler) configureYamllsEnabledGlob(config util.YamllsConfiguration) {
	globObject := config.GetEnabledForFilesGlobObject()
	h.yamllsConnector.EnabledForFilesGlobObject = globObject
}

func (h *TemplateHandler) configureYamlls(ctx context.Context, config util.YamllsConfiguration) {
	if config.Enabled {
		h.setYamllsConnector(yamlls.NewConnector(ctx, config, h.client, h.documents, &yamlls.DefaultCustomHandler))
		h.configureYamllsEnabledGlob(config)
		err := h.yamllsConnector.CallInitialize(ctx, h.chartStore.RootURI)
		if err != nil {
			logger.Error("Error initializing yamlls", err)
		}

		h.yamllsConnector.InitiallySyncOpenTemplateDocuments(h.documents.GetAllTemplateDocs())
	}
}
