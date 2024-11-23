package lsp

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
)

func TestTemplateContextVisitor(t *testing.T) {
	tC := struct {
		desc        string
		template    string
		nodeContent string
		expected    TemplateContext
	}{
		template: "{{ .Values.ingress.host }}",
	}

	s := &SymbolTable{
		contexts:         map[string][]sitter.Range{},
		contextsReversed: map[sitter.Range]TemplateContext{},
	}

	ast := ParseAst(nil, tC.template)
	rootNode := ast.RootNode()
	v := Visitors{
		symbolTable: s,
		visitors: []Visitor{
			NewTemplateContextVisitor(s, []byte(tC.template)),
		},
	}

	v.visitNodesRecursiveWithScopeShift(rootNode)

	keys := make([]sitter.Range, 0, len(s.contextsReversed))
	for r := range s.contextsReversed {
		keys = append(keys, r)
	}

	t.Log("Keys", keys)

	assert.Len(t, s.GetTemplateContextRanges([]string{"Values"}), 1)
	assert.Len(t, s.GetTemplateContextRanges([]string{"Values", "ingress"}), 1)
	assert.Len(t, s.GetTemplateContextRanges([]string{"Values", "ingress", "host"}), 1)

	templateContext, err := s.GetTemplateContext(sitter.Range{
		StartPoint: sitter.Point{Column: 10},
		EndPoint:   sitter.Point{Column: 11},
		StartByte:  10,
		EndByte:    11,
	})

	assert.NoError(t, err)
	assert.NotNil(t, templateContext)
	assert.Equal(t, TemplateContext{"Values", ""}, templateContext)
}

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
