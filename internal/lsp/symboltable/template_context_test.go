package symboltable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateContextFromYAMLPath(t *testing.T) {
	type testCase struct {
		yamlPath                string
		expectedTemplateContext TemplateContext
	}

	testCases := []testCase{
		{
			yamlPath:                "$.Values.test",
			expectedTemplateContext: TemplateContext{"Values", "test"},
		},
		{
			yamlPath:                "$.Values[0].test",
			expectedTemplateContext: TemplateContext{"Values[]", "test"},
		},
		{
			yamlPath:                "$.Values[0].test",
			expectedTemplateContext: TemplateContext{"Values[]", "test"},
		},
		{
			yamlPath:                "$.Values[0].test[999999999]",
			expectedTemplateContext: TemplateContext{"Values[]", "test[]"},
		},
		{
			yamlPath:                "something.in.there",
			expectedTemplateContext: TemplateContext{"something", "in", "there"},
		},
		{
			yamlPath:                ".something.in.there",
			expectedTemplateContext: TemplateContext{"", "something", "in", "there"},
		},
	}

	for _, v := range testCases {
		t.Run(v.yamlPath, func(t *testing.T) {
			result := TemplateContextFromYAMLPath(v.yamlPath)
			assert.Equal(t, v.expectedTemplateContext, result)
		})
	}
}
