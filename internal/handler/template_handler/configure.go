package templatehandler

import (
	"context"

	"github.com/gobwas/glob"
	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/util"
)

func (h *TemplateHandler) Configure(ctx context.Context, helmlsConfig util.HelmlsConfiguration) {
	h.configureYamlls(ctx, helmlsConfig)
}

func (h *TemplateHandler) configureYamlsEnabledGlob(helmlsConfig util.HelmlsConfiguration) {
	globObject, err := glob.Compile(helmlsConfig.YamllsConfiguration.EnabledForFilesGlob)
	if err != nil {
		logger.Error("Error compiling glob for yamlls EnabledForFilesGlob", err)
		globObject = util.DefaultConfig.YamllsConfiguration.EnabledForFilesGlobObject
	}
	h.yamllsConnector.EnabledForFilesGlobObject = globObject
}

func (h *TemplateHandler) configureYamlls(ctx context.Context, helmlsConfig util.HelmlsConfiguration) {
	config := helmlsConfig
	if config.YamllsConfiguration.Enabled {
		h.configureYamlsEnabledGlob(helmlsConfig)
		h.setYamllsConnector(yamlls.NewConnector(ctx, config.YamllsConfiguration, h.client, h.documents, &yamlls.DefaultCustomHandler))
		err := h.yamllsConnector.CallInitialize(ctx, h.chartStore.RootURI)
		if err != nil {
			logger.Error("Error initializing yamlls", err)
		}

		h.yamllsConnector.InitiallySyncOpenTemplateDocuments(h.documents.GetAllTemplateDocs())
	}
}
