package util

import sitter "github.com/smacker/go-tree-sitter"

func RangeContainsRange(surrounding, including sitter.Range) bool {
	return surrounding.StartByte <= including.StartByte && surrounding.EndByte >= including.EndByte
}

func NodeToRange(node *sitter.Node) sitter.Range {
	return sitter.Range{
		StartPoint: node.StartPoint(),
		EndPoint:   node.EndPoint(),
		StartByte:  node.StartByte(),
		EndByte:    node.EndByte(),
	}
}
