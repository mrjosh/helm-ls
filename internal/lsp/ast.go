package lsp

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
)

func ParseAst(oldTree *sitter.Tree, content string) *sitter.Tree {
	parser := sitter.NewParser()
	parser.SetLanguage(gotemplate.GetLanguage())
	tree, _ := parser.ParseCtx(context.Background(), oldTree, []byte(content))
	return tree
}

func NodeAtPosition(tree *sitter.Tree, position lsp.Position) *sitter.Node {
	start := sitter.Point{Row: position.Line, Column: position.Character}
	return tree.RootNode().NamedDescendantForPointRange(start, start)
}

func FindDirectChildNodeByStart(currentNode *sitter.Node, pointToLookUp sitter.Point) *sitter.Node {
	for i := 0; i < int(currentNode.ChildCount()); i++ {
		child := currentNode.Child(i)
		if child.StartPoint().Column == pointToLookUp.Column && child.StartPoint().Row == pointToLookUp.Row {
			return child
		}
	}
	return currentNode
}

func FindRelevantChildNode(currentNode *sitter.Node, pointToLookUp sitter.Point) *sitter.Node {
	for i := 0; i < int(currentNode.ChildCount()); i++ {
		child := currentNode.Child(i)
		if child == nil {
			continue
		}
		if isPointLargerOrEq(pointToLookUp, child.StartPoint()) && isPointLargerOrEq(child.EndPoint(), pointToLookUp) {
			return FindRelevantChildNode(child, pointToLookUp)
		}
	}
	return currentNode
}

func FindRelevantChildNodeCompletion(currentNode *sitter.Node, pointToLookUp sitter.Point) *sitter.Node {
	childCount := int(currentNode.ChildCount())
	for i := childCount - 1; i >= 0; i-- {
		child := currentNode.Child(i)
		if child == nil {
			continue
		}
		if isPointLargerOrEq(pointToLookUp, child.StartPoint()) && isPointLargerOrEq(child.EndPoint(), pointToLookUp) {
			return FindRelevantChildNodeCompletion(child, pointToLookUp)
		}
	}
	if currentNode.Type() == " " {
		return FindRelevantChildNodeCompletion(currentNode.Parent(), sitter.Point{
			Row:    pointToLookUp.Row,
			Column: pointToLookUp.Column - 1,
		})
	}
	return currentNode
}

func isPointLargerOrEq(a sitter.Point, b sitter.Point) bool {
	if a.Row == b.Row {
		return a.Column >= b.Column
	}
	return a.Row > b.Row
}

func (d *Document) ApplyChangesToAst(newContent string) {
	d.Ast = ParseAst(nil, newContent)
}

func GetLspRangeForNode(node *sitter.Node) lsp.Range {
	start := node.StartPoint()
	end := node.EndPoint()

	return lsp.Range{
		Start: lsp.Position{
			Line:      start.Row,
			Character: start.Column,
		},
		End: lsp.Position{
			Line:      end.Row,
			Character: end.Column,
		},
	}
}

func GetSitterPointForLspPos(pos lsp.Position) sitter.Point {
	return sitter.Point{
		Row:    pos.Line,
		Column: pos.Character,
	}
}
