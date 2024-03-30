package lsp

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
)

func Test_getTextNodeRanges(t *testing.T) {
	type args struct {
		gotemplateString string
	}
	tests := []struct {
		name string
		args args
		want []sitter.Range
	}{
		{
			name: "no text nodes",
			args: args{
				"{{ with .Values }}{{ .test }}{{ end }}",
			},
			want: []sitter.Range{},
		},
		{
			name: "simple text node",
			args: args{
				"a: {{ .test }}",
			},
			want: []sitter.Range{
				{
					StartPoint: sitter.Point{Row: 0, Column: 0},
					EndPoint:   sitter.Point{Row: 0, Column: 2},
					StartByte:  0,
					EndByte:    2,
				},
			},
		},
		{
			name: "to simple text nodes",
			args: args{
				`
a: {{ .test }}
b: not`,
			},
			want: []sitter.Range{
				{StartPoint: sitter.Point{
					Row: 0x1, Column: 0x0,
				}, EndPoint: sitter.Point{
					Row: 0x1, Column: 0x2,
				}, StartByte: 0x1, EndByte: 0x3},
				{
					StartPoint: sitter.Point{
						Row:    0x2,
						Column: 0x0,
					},
					EndPoint: sitter.Point{
						Row:    0x2,
						Column: 0x2,
					},
					StartByte: 0x10,
					EndByte:   0x12,
				},
				{
					StartPoint: sitter.Point{
						Row:    0x2,
						Column: 0x2,
					}, EndPoint: sitter.Point{
						Row:    0x2,
						Column: 0x6,
					},
					StartByte: 0x12,
					EndByte:   0x16,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTextNodeRanges(ParseAst(nil, tt.args.gotemplateString).RootNode())
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTrimTemplate(t *testing.T) {
	tests := []struct {
		documentText string
		trimmedText  string
	}{
		{
			documentText: `
{{ .Values.global. }}
yaml: test

{{block "name"}} T1 {{end}}
`,
			trimmedText: `
                     
yaml: test

                 T1        
`,
		},
		{
			documentText: `
{{ .Values.global. }}
yaml: test

{{block "name"}} T1 {{end}}
`,

			trimmedText: `
                     
yaml: test

                 T1        
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.documentText, func(t *testing.T) {
			gotemplateTree := ParseAst(nil, tt.documentText)
			got := TrimTemplate(gotemplateTree, tt.documentText)
			assert.Equal(t, tt.trimmedText, got)
		})
	}
}
