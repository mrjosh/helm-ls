package lsp

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

func TestIsInElseBranch(t *testing.T) {
	template := `{{if pipeline}} t1 {{ else if pipeline }} t2 {{ else }} t3 {{ end }}`
	var ast = ParseAst(template)
	// (template
	//     (if_action
	//         (function_call (identifier))
	//         (text)
	//         (function_call (identifier))
	//         (text)
	//         (text)))

	t1_start := sitter.Point{Row: 0, Column: 16}
	t1 := ast.RootNode().NamedDescendantForPointRange(t1_start, t1_start)
	t2_start := sitter.Point{Row: 0, Column: 42}
	t2 := ast.RootNode().NamedDescendantForPointRange(t2_start, t2_start)
	t3_start := sitter.Point{Row: 0, Column: 56}
	t3 := ast.RootNode().NamedDescendantForPointRange(t3_start, t3_start)

	if (t1.Content([]byte(template))) != " t1 " || (t2.Content([]byte(template))) != " t2 " || (t3.Content([]byte(template))) != " t3 " {
		t.Errorf("Nodes were not correclty selected")
	}

	if IsInElseBranch(t1) {
		t.Errorf("t1 was incorrectly identified as in else branch")
	}
	if !IsInElseBranch(t2) {
		t.Errorf("t2 was incorrectly identified as not in else branch")
	}
	if !IsInElseBranch(t3) {
		t.Errorf("t3 was incorrectly identified as not in else branch")
	}

}
