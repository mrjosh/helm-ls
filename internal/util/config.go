package util

type HelmlsConfiguration struct {
	YamllsConfiguration YamllsConfiguration `json:"yamlls,omitempty"`
	LogLevel            string              `json:"logLevel,omitempty"`
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
	YamllsConfiguration: YamllsConfiguration{
		Enabled:                 true,
		Path:                    "yaml-language-server",
		DiagnosticsLimit:        50,
		ShowDiagnosticsDirectly: false,
		YamllsSettings:          DefaultYamllsSettings,
	},
}

type YamllsSettings struct {
	Schemas    map[string]string `json:"schemas"`
	Completion bool              `json:"completion"`
	Hover      bool              `json:"hover"`
}

var DefaultYamllsSettings = YamllsSettings{
	Schemas:    map[string]string{"kubernetes": "**"},
	Completion: true,
	Hover:      true,
}
