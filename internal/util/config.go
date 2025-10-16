package util

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"

	"github.com/gobwas/glob"
)

type HelmlsConfiguration struct {
	YamllsConfiguration YamllsConfiguration `json:"yamlls,omitempty"`
	ValuesFilesConfig   ValuesFilesConfig   `json:"valuesFiles,omitempty"`
	HelmLintConfig      HelmLintConfig      `json:"helmLint,omitempty"`
	LogLevel            string              `json:"logLevel,omitempty"`
}

type ValuesFilesConfig struct {
	MainValuesFileName               string `json:"mainValuesFile,omitempty"`
	LintOverlayValuesFileName        string `json:"lintOverlayValuesFile,omitempty"`
	AdditionalValuesFilesGlobPattern string `json:"additionalValuesFilesGlobPattern,omitempty"`
}

type HelmLintConfig struct {
	Enabled         bool     `json:"enabled,omitempty"`
	IgnoredMessages []string `json:"ignoredMessages,omitempty"`
}

// YamllsPath can be either a string (backwards compatibility) or an array of strings
type YamllsPath []string

func (p *YamllsPath) UnmarshalJSON(data []byte) error {
	// Try unmarshaling as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*p = YamllsPath{s}
		return nil
	}

	// Try unmarshaling as []string
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*p = YamllsPath(arr)
		return nil
	}

	return &json.UnmarshalTypeError{
		Value: string(data),
		Type:  reflect.TypeOf([]string{}),
	}
}

func (p *YamllsPath) GetExecutable() string {
	if len(*p) > 0 {
		return (*p)[0]
	}
	return "yaml-language-server"
}

func (p *YamllsPath) GetArgs() []string {
	requiredArgs := []string{"--stdio"}
	if len(*p) > 1 {
		if (*p)[len(*p)-1] == "--stdio" {
			return (*p)[1:]
		} else {
			return append((*p)[1:], requiredArgs...)
		}
	}
	return requiredArgs
}

const YAMLLS_PATH_ENV_VAR = "YAMLLS_PATH"

type YamllsConfiguration struct {
	Enabled                   bool   `json:"enabled,omitempty"`
	EnabledForFilesGlob       string `json:"enabledForFilesGlob,omitempty"`
	EnabledForFilesGlobObject glob.Glob
	Path                      YamllsPath `json:"path,omitempty"`
	InitTimeoutSeconds        int        `json:"initTimeoutSeconds,omitempty"`
	DiagnosticsEnabled        bool       `json:"diagnosticsEnabled,omitempty"`
	// max diagnostics from yamlls that are shown for a single file
	DiagnosticsLimit int `json:"diagnosticsLimit,omitempty"`
	// if set to false diagnostics will only be shown after saving the file
	// otherwise writing a template will cause a lot of diagnostics to be shown because
	// the structure of the document is broken during typing
	ShowDiagnosticsDirectly bool `json:"showDiagnosticsDirectly,omitempty"`
	YamllsSettings          any  `json:"config,omitempty"`
}

func (y *YamllsConfiguration) CompileEnabledForFilesGlobObject() {
	globObject, err := glob.Compile(y.EnabledForFilesGlob)
	if err != nil {
		logger.Error("Error compiling glob for yamlls EnabledForFilesGlob", err)
		globObject = DefaultConfig.YamllsConfiguration.EnabledForFilesGlobObject
	}
	y.EnabledForFilesGlobObject = globObject
}

func (y *YamllsConfiguration) UpdatePathFromEnv() {
	if raw, ok := os.LookupEnv(YAMLLS_PATH_ENV_VAR); ok {
		var cleaned []string

		// Try to parse as JSON array first
		if err := json.Unmarshal([]byte(raw), &cleaned); err == nil {
			// Successfully parsed as JSON array
			if len(cleaned) > 0 {
				y.Path = cleaned
			}
		} else {
			// Fall back to comma-separated format
			parts := strings.Split(raw, ",")
			cleaned := make([]string, 0, len(parts))
			for _, p := range parts {
				if s := strings.TrimSpace(p); s != "" {
					cleaned = append(cleaned, s)
				}
			}
			if len(cleaned) > 0 {
				y.Path = cleaned
			}
		}
	}
}

var DefaultConfig = HelmlsConfiguration{
	LogLevel: "info",
	ValuesFilesConfig: ValuesFilesConfig{
		MainValuesFileName:               "values.yaml",
		LintOverlayValuesFileName:        "values.lint.yaml",
		AdditionalValuesFilesGlobPattern: "values*.yaml",
	},
	HelmLintConfig: HelmLintConfig{
		Enabled:         true,
		IgnoredMessages: []string{},
	},
	YamllsConfiguration: YamllsConfiguration{
		Enabled:                   true,
		EnabledForFilesGlob:       "*.{yaml,yml}",
		EnabledForFilesGlobObject: glob.MustCompile("*.{yaml,yml}"),
		Path:                      YamllsPath{"yaml-language-server"},
		InitTimeoutSeconds:        3,
		DiagnosticsEnabled:        true,
		DiagnosticsLimit:          50,
		ShowDiagnosticsDirectly:   false,
		YamllsSettings:            DefaultYamllsSettings,
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
