package util

import sitter "github.com/smacker/go-tree-sitter"

func RangeContainsRange(surrounding, including sitter.Range) bool {
	return surrounding.StartByte <= including.StartByte && surrounding.EndByte >= including.EndByte
}
