package util

type HelmlsConfiguration struct {
	YamllsConfiguration YamllsConfiguration `json:"yamlls,omitempty"`
	LogLevel            string              `json:"logLevel,omitempty"`
}

type YamllsConfiguration struct {
	Enabled bool   `json:"enabled,omitempty"`
	Path    string `json:"path,omitempty"`
	// max diagnostics from yamlls that are shown for a single file
	DiagnosticsLimit int            `json:"diagnosticsLimit,omitempty"`
	YamllsSettings   YamllsSettings `json:"config,omitempty"`
}

var DefaultConfig = HelmlsConfiguration{
	LogLevel: "info",
	YamllsConfiguration: YamllsConfiguration{
		Enabled:          true,
		Path:             "yaml-language-server",
		DiagnosticsLimit: 50,
		YamllsSettings:   DefaultYamllsSettings,
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
