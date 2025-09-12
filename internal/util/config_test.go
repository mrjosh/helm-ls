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
	// Set up the environment variable for the test
	os.Setenv(YAMLLS_PATH_ENV_VAR, "/test/path")

	config := &YamllsConfiguration{}
	config.UpdatePathFromEnv()

	// Use assert to check the expected value
	assert.Equal(t, YamllsPath{"/test/path"}, config.Path, "The path should be updated to the environment variable value")

	// Clean up the environment variable
	os.Unsetenv(YAMLLS_PATH_ENV_VAR)
}

func TestUpdatePathFromEnvSplit(t *testing.T) {
	// Set up the environment variable for the test
	os.Setenv(YAMLLS_PATH_ENV_VAR, "/test/path, yamlls.js")

	config := &YamllsConfiguration{}
	config.UpdatePathFromEnv()

	// Use assert to check the expected value
	assert.Equal(t, YamllsPath{"/test/path", " yamlls.js"}, config.Path, "The path should be updated to the environment variable value")

	// Clean up the environment variable
	os.Unsetenv(YAMLLS_PATH_ENV_VAR)
}

func TestUpdatePathFromEnv_Empty(t *testing.T) {
	// Ensure the environment variable is not set
	os.Unsetenv(YAMLLS_PATH_ENV_VAR)

	config := &YamllsConfiguration{}
	config.UpdatePathFromEnv()

	// Use assert to check the expected value
	assert.Empty(t, config.Path, "The path should be empty when the environment variable is not set")
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
