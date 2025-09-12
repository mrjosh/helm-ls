package util

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYamllsPath_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError error
		expected      YamllsPath
	}{
		{
			name:     "single string",
			input:    `"yaml-language-server"`,
			expected: YamllsPath{"yaml-language-server"},
		},
		{
			name:     "string array",
			input:    `["yaml-language-server", "--foo", "--bar"]`,
			expected: YamllsPath{"yaml-language-server", "--foo", "--bar"},
		},
		{
			name:     "string array",
			input:    `["node","yaml-language-server.js", "--foo", "--bar"]`,
			expected: YamllsPath{"node", "yaml-language-server.js", "--foo", "--bar"},
		},
		{
			name:          "error",
			input:         `{}`,
			expectedError: &json.UnmarshalTypeError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path YamllsPath
			err := json.Unmarshal([]byte(tt.input), &path)

			if tt.expectedError != nil {
				var ute *json.UnmarshalTypeError
				assert.ErrorAs(t, err, &ute)
				return
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, path)
		})
	}
}

func TestUpdatePathFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected YamllsPath
	}{
		{
			name:     "Single Path",
			envValue: "/test/path",
			expected: YamllsPath{"/test/path"},
		},
		{
			name:     "Multiple Paths",
			envValue: "/test/path,yamlls.js",
			expected: YamllsPath{"/test/path", "yamlls.js"},
		},
		{
			name:     "Empty Path",
			envValue: "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(YAMLLS_PATH_ENV_VAR, tt.envValue)
			} else {
				os.Unsetenv(YAMLLS_PATH_ENV_VAR)
			}

			config := &YamllsConfiguration{}
			config.UpdatePathFromEnv()

			assert.Equal(t, tt.expected, config.Path, "The path should be updated to the environment variable value")

			os.Unsetenv(YAMLLS_PATH_ENV_VAR)
		})
	}
}

func TestYamllsPath_GetExecutable(t *testing.T) {
	tests := []struct {
		name     string
		path     YamllsPath
		expected string
	}{
		{"empty path", YamllsPath{}, "yaml-language-server"},
		{"non-empty path", YamllsPath{"custom-server", "--flag"}, "custom-server"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.path.GetExecutable())
		})
	}
}

func TestYamllsPath_GetArgs(t *testing.T) {
	tests := []struct {
		name     string
		path     YamllsPath
		expected []string
	}{
		{
			"empty path",
			YamllsPath{},
			[]string{"--stdio"},
		},
		{
			"only executable",
			YamllsPath{"custom-server"},
			[]string{"--stdio"},
		},
		{
			"executable with args (last is --stdio)",
			YamllsPath{"custom-server", "--foo", "--stdio"},
			[]string{"--foo", "--stdio"},
		},
		{
			"executable with args (last not --stdio)",
			YamllsPath{"custom-server", "--foo"},
			[]string{"--foo", "--stdio"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.path.GetArgs())
		})
	}
}
