package yamlls

import (
	sitter "github.com/smacker/go-tree-sitter"
)

func trimTemplateForYamllsFromAst(ast *sitter.Tree, text string) string {
	var result = []byte(text)
	prettyPrintNode(ast.RootNode(), []byte(text), result)
	return string(result)
}

func prettyPrintNode(node *sitter.Node, previous []byte, result []byte) {
	var childCount = node.ChildCount()

	switch node.Type() {
	case "if_action":
		curser := sitter.NewTreeCursor(node)
		curser.GoToFirstChild()
		for curser.GoToNextSibling() {
			if curser.CurrentFieldName() == "condition" {
				earaseTemplate(curser.CurrentNode(), previous, result)
				earaseTemplate(curser.CurrentNode().NextSibling(), previous, result)
				continue
			}
			switch curser.CurrentNode().Type() {
			case "if", "else if":
				earaseTemplate(curser.CurrentNode(), previous, result)
				earaseTemplate(curser.CurrentNode().PrevSibling(), previous, result)
			case "end":
				earaseTemplate(curser.CurrentNode(), previous, result)
				earaseTemplate(curser.CurrentNode().PrevSibling(), previous, result)
				earaseTemplate(curser.CurrentNode().NextSibling(), previous, result)
			default:
				prettyPrintNode(curser.CurrentNode(), previous, result)
			}
		}
		curser.Close()

	case "block_action", "with_action", "range_action":
		for i := 0; i < int(childCount); i++ {
			child := node.Child(i)
			switch child.Type() {
			case
				"if",
				"selector_expression",
				"else",
				"range",
				"function_call",
				"with",
				"define",
				"{{",
				"{{-",
				"}}",
				"-}}",
				"end",
				"interpreted_string_literal",
				"block":
				earaseTemplate(child, previous, result)
			default:
				prettyPrintNode(child, previous, result)
			}
		}
	case "define_action":
		earaseTemplate(node, previous, result)
	case "comment", "variable_definition":
		earaseTemplate(node, previous, result)
		var prevSibling, nextSibling = node.PrevSibling(), node.NextSibling()
		if prevSibling != nil {
			earaseTemplate(prevSibling, previous, result)
		}
		if nextSibling != nil {
			earaseTemplate(nextSibling, previous, result)
		}
	default:
		for i := 0; i < int(childCount); i++ {
			prettyPrintNode(node.Child(i), previous, result)
		}
	}
}

func earaseTemplate(node *sitter.Node, previous []byte, result []byte) {
	logger.Println("Content that is earased", node.Content(previous))
	for i := range []byte(node.Content(previous)) {
		result[int(node.StartByte())+i] = byte(' ')
	}
}
