package gotemplate

//#include "tree_sitter/parser.h"
//TSLanguage *tree_sitter_gotmpl();
import "C"
import (
	sitter "github.com/smacker/go-tree-sitter"
	"unsafe"
)

func GetLanguage() *sitter.Language {
	ptr := unsafe.Pointer(C.tree_sitter_gotmpl())
	return sitter.NewLanguage(ptr)
}
