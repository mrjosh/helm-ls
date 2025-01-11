package templatehandler

import (
	"context"

	"github.com/gobwas/glob"
	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/util"
)

func (h *TemplateHandler) Configure(ctx context.Context, helmlsConfig util.HelmlsConfiguration) {
	h.configureYamlls(ctx, helmlsConfig.YamllsConfiguration)
}

func (h *TemplateHandler) configureYamlsEnabledGlob(config util.YamllsConfiguration) {
	globObject, err := glob.Compile(config.EnabledForFilesGlob)
	if err != nil {
		logger.Error("Error compiling glob for yamlls EnabledForFilesGlob", err)
		globObject = util.DefaultConfig.YamllsConfiguration.EnabledForFilesGlobObject
	}
	h.yamllsConnector.EnabledForFilesGlobObject = globObject
}

func (h *TemplateHandler) configureYamlls(ctx context.Context, config util.YamllsConfiguration) {
	if config.Enabled {
		h.configureYamlsEnabledGlob(config)
		h.setYamllsConnector(yamlls.NewConnector(ctx, config, h.client, h.documents, &yamlls.DefaultCustomHandler))
		err := h.yamllsConnector.CallInitialize(ctx, h.chartStore.RootURI)
		if err != nil {
			logger.Error("Error initializing yamlls", err)
		}

		h.yamllsConnector.InitiallySyncOpenTemplateDocuments(h.documents.GetAllTemplateDocs())
	}
}
