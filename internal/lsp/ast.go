package lsp

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
)

func ParseAst(content string) *sitter.Tree {
	parser := sitter.NewParser()
	parser.SetLanguage(gotemplate.GetLanguage())
	tree, _ := parser.ParseCtx(context.Background(), nil, []byte(content))
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
		if isPointLargerOrEq(pointToLookUp, child.StartPoint()) && isPointLargerOrEq(child.EndPoint(), pointToLookUp) {
			return FindRelevantChildNode(child, pointToLookUp)
		}
	}
	return currentNode
}

func isPointLargerOrEq(a sitter.Point, b sitter.Point) bool {
	if a.Row == b.Row {
		return a.Column >= b.Column
	}
	return a.Row > b.Row
}

func GetFieldIdentifierPath(node *sitter.Node, doc *Document) (path string) {
	path = buildFieldIdentifierPath(node, doc)
	logger.Debug("buildFieldIdentifierPath:", path)
	return path
}

func buildFieldIdentifierPath(node *sitter.Node, doc *Document) string {

	prepend := node.PrevNamedSibling()

	currentPath := node.Content([]byte(doc.Content))
	if prepend != nil {
		nodeContent := node.Content([]byte(doc.Content))
		if nodeContent == "." {
			nodeContent = ""
		}
		currentPath = prepend.Content([]byte(doc.Content)) + "." + nodeContent
	}

	if currentPath[0:1] == "$" {
		return currentPath
	}

	if currentPath[0:1] != "." {
		currentPath = "." + currentPath
	}

	return TraverseIdentifierPathUp(node, doc) + currentPath
}

func TraverseIdentifierPathUp(node *sitter.Node, doc *Document) string {
	parent := node.Parent()

	if parent == nil {
		return ""
	}

	switch parent.Type() {
	case "range_action":
		if node.PrevNamedSibling() == nil {
			return TraverseIdentifierPathUp(parent, doc)
		}
		logger.Debug("Range action found")
		return TraverseIdentifierPathUp(parent, doc) + parent.NamedChild(0).Content([]byte(doc.Content)) + "[0]"
	case "with_action":
		if node.PrevNamedSibling() == nil {
			return TraverseIdentifierPathUp(parent, doc)
		}
		logger.Debug("With action found")
		return TraverseIdentifierPathUp(parent, doc) + parent.NamedChild(0).Content([]byte(doc.Content))
	}
	return TraverseIdentifierPathUp(parent, doc)
}

func (d *Document) ApplyChangesToAst(newContent string) {
	d.Ast = ParseAst(newContent)
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
