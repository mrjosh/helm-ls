package util

type HelmlsConfiguration struct {
	YamllsConfiguration YamllsConfiguration `json:"yamlls,omitempty"`
	ValuesFilesConfig   ValuesFilesConfig   `json:"valuesFiles,omitempty"`
	LogLevel            string              `json:"logLevel,omitempty"`
}

type ValuesFilesConfig struct {
	MainValuesFileName               string `json:"mainValuesFile,omitempty"`
	LintOverlayValuesFileName        string `json:"lintOverlayValuesFile,omitempty"`
	AdditionalValuesFilesGlobPattern string `json:"additionalValuesFilesGlobPattern,omitempty"`
}

type YamllsConfiguration struct {
	Enabled bool   `json:"enabled,omitempty"`
	Path    string `json:"path,omitempty"`
	// max diagnostics from yamlls that are shown for a single file
	DiagnosticsLimit int `json:"diagnosticsLimit,omitempty"`
	// if set to false diagnostics will only be shown after saving the file
	// otherwise writing a template will cause a lot of diagnostics to be shown because
	// the structure of the document is broken during typing
	ShowDiagnosticsDirectly bool        `json:"showDiagnosticsDirectly,omitempty"`
	YamllsSettings          interface{} `json:"config,omitempty"`
}

var DefaultConfig = HelmlsConfiguration{
	LogLevel: "info",
	ValuesFilesConfig: ValuesFilesConfig{
		MainValuesFileName:               "values.yaml",
		LintOverlayValuesFileName:        "values.lint.yaml",
		AdditionalValuesFilesGlobPattern: "values*.yaml",
	},
	YamllsConfiguration: YamllsConfiguration{
		Enabled:                 true,
		Path:                    "yaml-language-server",
		DiagnosticsLimit:        50,
		ShowDiagnosticsDirectly: false,
		YamllsSettings:          DefaultYamllsSettings,
	},
}

type YamllsSchemaStoreSettings struct {
	Enable bool   `json:"enable"`
	URL    string `json:"url"`
}

type YamllsSettings struct {
	Schemas                   map[string]string         `json:"schemas"`
	Completion                bool                      `json:"completion"`
	Hover                     bool                      `json:"hover"`
	YamllsSchemaStoreSettings YamllsSchemaStoreSettings `json:"schemaStore"`
}

var DefaultYamllsSettings = YamllsSettings{
	Schemas:    map[string]string{"kubernetes": "templates/**"},
	Completion: true,
	Hover:      true,
	YamllsSchemaStoreSettings: YamllsSchemaStoreSettings{
		Enable: true,
		URL:    "https://www.schemastore.org/api/json/catalog.json",
	},
}
