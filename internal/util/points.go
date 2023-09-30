package util

import (
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
)

func PointToPosition(point sitter.Point) lsp.Position {
	return lsp.Position{Line: point.Row, Character: point.Column}
}

func PositionToPoint(position lsp.Position) sitter.Point {
	return sitter.Point{Row: position.Line, Column: position.Character}
}
