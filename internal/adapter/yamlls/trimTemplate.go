package yamlls

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

func trimTemplateForYamllsFromAst(ast *sitter.Tree, text string) string {
	var result = []byte(text)
	// logger.Println(ast.RootNode())
	prettyPrintNode(ast.RootNode(), []byte(text), result)
	return string(result)
}

func prettyPrintNode(node *sitter.Node, previous []byte, result []byte) {
	var childCount = node.ChildCount()

	switch node.Type() {
	case "if_action":
		for i := 0; i < int(childCount); i++ {
			logger.Debug("FieldName", node.FieldNameForChild(i))
			child := node.Child(i)
			if child.Type() == "end" {
				earaseTemplate(child, previous, result)
				earaseTemplate(child.NextSibling(), previous, result)
				earaseTemplate(child.PrevSibling(), previous, result)
				break
			} else {
				prettyPrintNode(child, previous, result)
			}
		}
		if_action_condition := node.ChildByFieldName("condition")
		earaseTemplate(if_action_condition, previous, result)
		earaseTemplate(if_action_condition.NextSibling(), previous, result)
		if if_action_condition.PrevSibling() != nil && if_action_condition.PrevSibling().Type() == "if" {
			earaseTemplate(if_action_condition.PrevSibling(), previous, result)
			earaseTemplate(if_action_condition.PrevSibling().PrevSibling(), previous, result)
		}
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
	case "function_call":
		x := node.Content(previous)
		if strings.HasPrefix(x, "include ") {
			println("Function call", x)
			parent := node.Parent()

			if parent.Type() == "chained_pipeline" {
				replaceWithString(parent, previous, result)
			} else {
				replaceWithString(node, previous, result)
			}
		}
	default:
		for i := 0; i < int(childCount); i++ {
			prettyPrintNode(node.Child(i), previous, result)
		}
	}
}

func earaseTemplate(node *sitter.Node, previous []byte, result []byte) {
	logger.Debug("Content that is earased", node.Content(previous))
	for i := range []byte(node.Content(previous)) {
		result[int(node.StartByte())+i] = byte(' ')
	}
}

func replaceWithString(node *sitter.Node, previous []byte, result []byte) {
	logger.Debug("Content that is earased", node.Content(previous))
	for i := range []byte(node.Content(previous)) {
		result[int(node.StartByte())+i] = byte(' ')
		// if i == 0 || i == -1+len([]byte(node.Content(previous))) {
		// 	result[int(node.StartByte())+i] = byte('"')
		// }
	}
	prevSibling, nextSibling := node.PrevSibling(), node.NextSibling()

	if prevSibling != nil && (prevSibling.Type() == "{{-" || prevSibling.Type() == "{{") {
		result[int(prevSibling.StartByte())] = byte('"')
	}

	if nextSibling != nil && (nextSibling.Type() == "-}}" || nextSibling.Type() == "}}") {
		result[int(nextSibling.EndByte())-1] = byte('"')
	}

}
