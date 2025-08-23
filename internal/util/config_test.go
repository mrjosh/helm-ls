package util

import (
	"encoding/json"
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
