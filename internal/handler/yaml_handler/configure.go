package yamlhandler

import (
	"context"

	"github.com/gobwas/glob"
	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/util"
)

func (h *YamlHandler) Configure(ctx context.Context, helmlsConfig util.HelmlsConfiguration) {
	h.configureYamlls(ctx, helmlsConfig)
}

func (h *YamlHandler) configureYamlsEnabledGlob(helmlsConfig util.HelmlsConfiguration) {
	globObject, err := glob.Compile(helmlsConfig.YamllsConfiguration.EnabledForFilesGlob)
	if err != nil {
		logger.Error("Error compiling glob for yamlls EnabledForFilesGlob", err)
		globObject = util.DefaultConfig.YamllsConfiguration.EnabledForFilesGlobObject
	}
	h.yamllsConnector.EnabledForFilesGlobObject = globObject
}

func (h *YamlHandler) configureYamlls(ctx context.Context, helmlsConfig util.HelmlsConfiguration) {
	config := helmlsConfig
	if config.YamllsConfiguration.Enabled {
		h.configureYamlsEnabledGlob(helmlsConfig)
		h.setYamllsConnector(
			yamlls.NewConnector(ctx,
				config.YamllsConfiguration,
				h.client,
				h.documents,
				h.CustomHandler))
		err := h.yamllsConnector.CallInitialize(ctx, h.chartStore.RootURI)
		if err != nil {
			logger.Error("Error initializing yamlls", err)
		}

		h.yamllsConnector.InitiallySyncOpenDocuments(h.documents.GetAllTemplateDocs())
	}
}
