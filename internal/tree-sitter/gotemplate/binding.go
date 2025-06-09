package gotemplate

import (
	tree_sitter_gotmpl "github.com/qvalentin/tree-sitter-go-template/bindings/go"
	sitter "github.com/smacker/go-tree-sitter"
)

func GetLanguage() *sitter.Language {
	return tree_sitter_gotmpl.GetLanguage()
}
