package util

import (
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func PointToPosition(point sitter.Point) lsp.Position {
	return lsp.Position{Line: point.Row, Character: point.Column}
}

func PositionToPoint(position lsp.Position) sitter.Point {
	return sitter.Point{Row: position.Line, Column: position.Character}
}

func RangeToLocation(URI uri.URI, range_ sitter.Range) lsp.Location {
	return lsp.Location{
		URI: URI,
		Range: lsp.Range{
			Start: PointToPosition(range_.StartPoint),
			End:   PointToPosition(range_.EndPoint),
		},
	}
}

func GetRangeForNode(node *sitter.Node) sitter.Range {
	return sitter.Range{
		StartPoint: node.StartPoint(),
		EndPoint:   node.EndPoint(),
		StartByte:  node.StartByte(),
		EndByte:    node.EndByte(),
	}
}
