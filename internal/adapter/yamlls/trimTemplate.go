package yamlls

import (
	sitter "github.com/smacker/go-tree-sitter"
)

func trimTemplateForYamllsFromAst(ast *sitter.Tree, text string) string {

	var result = []byte(text)
	logger.Println(ast.RootNode())
	prettyPrintNode(ast.RootNode(), []byte(text), result)
	return string(result)

}

func prettyPrintNode(node *sitter.Node, previous []byte, result []byte) {
	var childCount = node.ChildCount()

	switch node.Type() {
	case "if_action", "block_action", "with_action", "range_action":
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
	case "comment", "define_action":
		earaseTemplate(node, previous, result)
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
