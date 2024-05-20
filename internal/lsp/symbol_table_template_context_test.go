package lsp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContextForSelectorExpression(t *testing.T) {
	testCases := []struct {
		desc        string
		template    string
		nodeContent string
		expected    TemplateContext
	}{
		{
			desc:        "Selects simple selector expression correctly",
			template:    `{{ .Values.test }}`,
			nodeContent: ".Values.test",
			expected:    TemplateContext{"Values", "test"},
		},
		{
			desc:        "Selects unfinished selector expression correctly",
			template:    `{{ .Values.test. }}`,
			nodeContent: ".Values.test.",
			expected:    TemplateContext{"Values", "test"},
		},
		{
			desc:        "Selects selector expression with $ correctly",
			template:    `{{ $.Values.test }}`,
			nodeContent: "$.Values.test",
			expected:    TemplateContext{"$", "Values", "test"},
		},
		{
			desc:        "Selects unfinished selector expression with $ correctly",
			template:    `{{ $.Values.test. }}`,
			nodeContent: "$.Values.test.",
			expected:    TemplateContext{"$", "Values", "test"},
		},
		{
			desc:        "Selects selector expression with variable correctly",
			template:    `{{ $x.test }}`,
			nodeContent: "$x.test",
			expected:    TemplateContext{"$x", "test"},
		},
		{
			desc:        "Selects unfinished selector expression with variable correctly",
			template:    `{{ $x.test. }}`,
			nodeContent: "$x.test.",
			expected:    TemplateContext{"$x", "test"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ast := ParseAst(nil, tC.template)
			node := ast.RootNode().Child(1)

			assert.Equal(t, tC.nodeContent, node.Content([]byte(tC.template)))
			result := getContextForSelectorExpression(node, []byte(tC.template))

			assert.Equal(t, tC.expected, result)
		})
	}
}
