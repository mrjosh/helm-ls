package lsp

import (
	"errors"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
)

func createSymbolTableWithContexts(template string) *SymbolTable {
	ast := ParseAst(nil, template)
	rootNode := ast.RootNode()
	s := &SymbolTable{
		contexts:         map[string][]sitter.Range{},
		contextsReversed: map[sitter.Range]TemplateContext{},
	}
	v := Visitors{
		symbolTable: s,
		visitors: []Visitor{
			NewTemplateContextVisitor(s, []byte(template)),
		},
	}
	v.visitNodesRecursiveWithScopeShift(rootNode)
	return s
}

func TestTemplateContextVisitor(t *testing.T) {
	template := "{{ .Values.ingress.host }}"
	s := createSymbolTableWithContexts(template)
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

func TestGetTemplateContextFromSymbolTable(t *testing.T) {
	testCases := []struct {
		desc                  string
		templateWithRangeMark string
		expected              TemplateContext
		err                   error
	}{
		{
			desc:                  "Selects end of simple selector expression correctly",
			templateWithRangeMark: `{{ .Values.^test^ }}`,
			expected:              TemplateContext{"Values", "test"},
		},
		{
			desc:                  "Selects start of simple selector expression correctly",
			templateWithRangeMark: `{{ .^Values^.test }}`,
			expected:              TemplateContext{"Values"},
		},
		{
			desc:                  "Selects dot of simple selector expression correctly",
			templateWithRangeMark: `{{ .Values^.^test }}`,
			expected:              TemplateContext{"Values", ""},
		},
		{
			desc:                  "Selects starting dot of simple selector expression correctly",
			templateWithRangeMark: `{{ ^.^Values.test }}`,
			expected:              TemplateContext{""},
		},
		{
			desc:                  "Selects dot of unfinished selector expression correctly",
			templateWithRangeMark: `{{ .Values.test^.^ }}`,
			expected:              TemplateContext{"Values", "test", ""},
		},
		{
			desc:                  "Selects starting dot of unfinished selector expression correctly",
			templateWithRangeMark: `{{ ^.^Values.test. }}`,
			expected:              TemplateContext{""},
		},
		{
			desc:                  "Selects middle dot of unfinished selector expression correctly",
			templateWithRangeMark: `{{ .Values^.^test. }}`,
			expected:              TemplateContext{"Values", ""},
		},
		{
			desc:                  "Selects dot correctly",
			templateWithRangeMark: ` {{ ^.^ }}`, // This is not consistent with dot in selector expression, but it's o.k.
			expected:              TemplateContext{},
		},
		{
			desc:                  "Selects dot in with correctly",
			templateWithRangeMark: `{{ with .Values.test }} {{ ^.^ }} {{ end }}`,
			expected:              TemplateContext{"Values", "test"},
		},
		{
			desc:                  "Selects selector expression in with correctly",
			templateWithRangeMark: `{{ with .Values.test }} {{ .^something^ }} {{ end }}`,
			expected:              TemplateContext{"Values", "test", "something"},
		},
		{
			desc:                  "Selects unfinished selector expression in with correctly",
			templateWithRangeMark: `{{ with .Values.test }} {{ .something^.^ }} {{ end }}`,
			expected:              TemplateContext{"Values", "test", "something", ""},
		},
		{
			desc:                  "Selects start dot of unfinished selector expression in with correctly",
			templateWithRangeMark: `{{ with .Values.test }} {{ ^.^something. }} {{ end }}`,
			expected:              TemplateContext{"Values", "test", ""},
		},
		{
			desc:                  "Selects root selector expression in with correctly",
			templateWithRangeMark: `{{ with .Values.test }} {{ $.Values.something^.^ }} {{ end }}`,
			expected:              TemplateContext{"Values", "something", ""},
		},
		{
			desc:                  "Selects dot in range correctly",
			templateWithRangeMark: `{{ range .Values.test }} {{ ^.^ }} {{ end }}`,
			expected:              TemplateContext{"Values", "test[]"},
		},
		{
			desc:                  "Selects selector expression in range correctly",
			templateWithRangeMark: `{{ range .Values.test }} {{ .^something^ }} {{ end }}`,
			expected:              TemplateContext{"Values", "test[]", "something"},
		},
		{
			desc:                  "Selects unfinished selector expression in range correctly",
			templateWithRangeMark: `{{ range .Values.test }} {{ .something^.^ }} {{ end }}`,
			expected:              TemplateContext{"Values", "test[]", "something", ""},
		},
		{
			desc:                  "Selects root selector expression in range correctly",
			templateWithRangeMark: `{{ range .Values.test }} {{ $.Values.something^.^ }} {{ end }}`,
			expected:              TemplateContext{"Values", "something", ""},
		},
		{
			desc:                  "Selects variable selector expression in range correctly without variable visitor",
			templateWithRangeMark: `{{ $x := .Values }} {{ range .Values.test }} {{ $x.something^.^ }} {{ end }}`,
			expected:              TemplateContext{"$x", "something", ""},
			err:                   errors.New("variable $x not found"), // because we have no variable visitor in this test
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			expectedRange, template := getRangeForMarkedTestLine(tC.templateWithRangeMark)
			s := createSymbolTableWithContexts(template)
			result, err := s.GetTemplateContext(expectedRange)

			assert.Equal(t, tC.err, err)
			assert.Equal(t, tC.expected, result)
		})
	}
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
