// The content of the Snippet was taken from https://pkg.go.dev/text/template
// So the following license applies to them:

// Copyright (c) 2009 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//   - Redistributions of source code must retain the above copyright
//
// notice, this list of conditions and the following disclaimer.
//   - Redistributions in binary form must reproduce the above
//
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//   - Neither the name of Google Inc. nor the names of its
//
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
package handler

type HelmSnippet struct {
	Name    string
	Detail  string
	Doc     string
	Snippet string
	Filter  string
}

var (
	textSnippets = []HelmSnippet{
		{
			Name:    "comment",
			Detail:  "{{- /* a comment with white space trimmed from preceding and following text */ -}}",
			Doc:     "A comment; discarded. May contain newlines. Comments do not nest and must start and end at the delimiters, as shown here.",
			Snippet: "{{- /* $1 */ -}}",
		},
		{
			Name:    "{{ }}",
			Detail:  "template",
			Doc:     "",
			Snippet: "{{- $0 }}",
			Filter:  "{}", // TODO: is this useful?
		},
		{
			Name:    "if",
			Detail:  "{{if pipeline}} T1 {{end}}",
			Doc:     "If the value of the pipeline is empty, no output is generated; otherwise, T1 is executed. The empty values are false, 0, any nil pointer or interface value, and any array, slice, map, or string of length zero. Dot is unaffected.",
			Snippet: "{{- if $1 }}\n $0 \n{{- end }}",
		},
		{
			Name:    "if else",
			Detail:  "{{if pipeline}} T1 {{else}} T0 {{end}}",
			Doc:     "If the value of the pipeline is empty, T0 is executed; otherwise, T1 is executed. Dot is unaffected.",
			Snippet: "{{- if $1 }}\n $2 \n{{- else }}\n $0 \n{{- end }}",
		},
		{
			Name:    "if else if",
			Detail:  "{{if pipeline}} T1 {{else if pipeline}} T0 {{end}}",
			Doc:     "To simplify the appearance of if-else chains, the else action of an if may include another if directly; the effect is exactly the same as writing {{if pipeline}} T1 {{else}}{{if pipeline}} T0 {{end}}{{end}}",
			Snippet: "{{- if $1 }}\n $2 \n{{- else if $4 }}\n $0 \n{{- end }}",
		},
		{
			Name:    "range",
			Detail:  "{{range pipeline}} T1 {{end}}",
			Doc:     "The value of the pipeline must be an array, slice, map, or channel. If the value of the pipeline has length zero, nothing is output; otherwise, dot is set to the successive elements of the array, slice, or map and T1 is executed. If the value is a map and the keys are of basic type with a defined order, the elements will be visited in sorted key order.",
			Snippet: "{{- range $1 }}\n $0 \n{{- end }}",
		},
		{
			Name:    "range else",
			Detail:  "{{range pipeline}} T1 {{else}} T0 {{end}}",
			Doc:     "The value of the pipeline must be an array, slice, map, or channel. If the value of the pipeline has length zero, dot is unaffected and T0 is executed; otherwise, dot is set to the successive elements of the array, slice, or map and T1 is executed.",
			Snippet: "{{- range $1 }}\n $2 {{- else }}\n $0 \n{{- end }}",
		},
		{
			Name:    "block",
			Detail:  "{{block \"name\" pipeline}} T1 {{end}}",
			Doc:     "A block is shorthand for defining a template {{define \"name\"}} T1 {{end}} and then executing it in place {{template \"name\" pipeline}} The typical use is to define a set of root templates that are then customized by redefining the block templates within.",
			Snippet: "{{- block $1 }}\n $0 \n{{- end }}",
		},
		{
			Name:    "with",
			Detail:  "{{with pipeline}} T1 {{end}}",
			Doc:     "If the value of the pipeline is empty, no output is generated; otherwise, dot is set to the value of the pipeline and T1 is executed.",
			Snippet: "{{- with $1 }}\n $0 \n{{- end }}",
		},
		{
			Name:    "with else",
			Detail:  "{{with pipeline}} T1 {{else}} T0 {{end}}",
			Doc:     "If the value of the pipeline is empty, dot is unaffected and T0 is executed; otherwise, dot is set to the value of the pipeline and T1 is executed",
			Snippet: "{{- with $1 }}\n $2 {{- else }}\n $0 \n{{- end }}",
		},
	}
)
