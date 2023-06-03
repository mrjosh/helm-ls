package lsp

import (
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
)

func ParseAst(content string) *sitter.Tree {
	parser := sitter.NewParser()
	parser.SetLanguage(gotemplate.GetLanguage())
	return parser.Parse(nil, []byte(content))
}

func NodeAtPosition(tree *sitter.Tree, position lsp.Position) *sitter.Node {
	start := sitter.Point{Row: position.Line, Column: position.Character}
	return tree.RootNode().NamedDescendantForPointRange(start, start)
}

func GetFieldIdentifierPath(node *sitter.Node, doc *document) (path string) {
	path = buildFieldIdentifierPath(node, doc)
	logger.Println("buildFieldIdentifierPath:", path)
	return path
}

func buildFieldIdentifierPath(node *sitter.Node, doc *document) string {

	prepend := node.PrevNamedSibling()

	currentPath := node.Content([]byte(doc.Content))
	if prepend != nil {
		currentPath = prepend.Content([]byte(doc.Content)) + "." + node.Content([]byte(doc.Content))
	}

	if currentPath[0:1] == "$" {
		return currentPath
	}

	if currentPath[0:1] != "." {
		currentPath = "." + currentPath
	}

	return TraverseIdentifierPathUp(node, doc) + currentPath
}

func TraverseIdentifierPathUp(node *sitter.Node, doc *document) string {
	parent := node.Parent()

	if parent == nil {
		return ""
	}

	switch parent.Type() {
	case "range_action":
		if node.PrevNamedSibling() == nil {
			return TraverseIdentifierPathUp(parent, doc)
		}
		logger.Println("Range action found ")
		return TraverseIdentifierPathUp(parent, doc) + parent.NamedChild(0).Content([]byte(doc.Content)) + "[0]"
	case "with_action":
		if node.PrevNamedSibling() == nil {
			return TraverseIdentifierPathUp(parent, doc)
		}
		logger.Println("With action found")
		return TraverseIdentifierPathUp(parent, doc) + parent.NamedChild(0).Content([]byte(doc.Content))
	}
	return TraverseIdentifierPathUp(parent, doc)
}

func (d *document) ApplyChangesToAst(newContent string) {
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
