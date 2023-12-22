package lsp

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

func TestIsInElseBranch(t *testing.T) {
	template := `{{if pipeline}} t1 {{ else if pipeline }} t2 {{ else if pipeline2 }} t3 {{ else }} t4 {{ end }}`
	var ast = ParseAst(template)
	// (template [0, 0] - [1, 0]
	//   (if_action [0, 0] - [0, 95]
	//     condition: (function_call [0, 5] - [0, 13]
	//       function: (identifier [0, 5] - [0, 13]))
	//     consequence: (text [0, 15] - [0, 18])
	//     condition: (function_call [0, 30] - [0, 38]
	//       function: (identifier [0, 30] - [0, 38]))
	//     option: (text [0, 41] - [0, 44])
	//     condition: (function_call [0, 56] - [0, 65]
	//       function: (identifier [0, 56] - [0, 65]))
	//     option: (text [0, 68] - [0, 71])
	//     alternative: (text [0, 82] - [0, 85])))

	logger.Println("RootNode:", ast.RootNode().String())

	t1_start := sitter.Point{Row: 0, Column: 16}
	t1 := ast.RootNode().NamedDescendantForPointRange(t1_start, t1_start)
	t2_start := sitter.Point{Row: 0, Column: 42}
	t2 := ast.RootNode().NamedDescendantForPointRange(t2_start, t2_start)
	t3_start := sitter.Point{Row: 0, Column: 69}
	t3 := ast.RootNode().NamedDescendantForPointRange(t3_start, t3_start)
	t4_start := sitter.Point{Row: 0, Column: 83}
	t4 := ast.RootNode().NamedDescendantForPointRange(t4_start, t4_start)
	t1Content := t1.Content([]byte(template))
	t2Content := t2.Content([]byte(template))
	t3Content := t3.Content([]byte(template))
	t4Content := t4.Content([]byte(template))
	if (t1Content != " t1") || t2Content != " t2" || t3Content != " t3" || t4Content != " t4" {
		t.Errorf("Nodes were not correctly selected")
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
	if !IsInElseBranch(t4) {
		t.Errorf("t4 was incorrectly identified as not in else branch")
	}

}
