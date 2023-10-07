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
		for i := 0; i < int(childCount); i++ {
			logger.Debug("FieldName", node.FieldNameForChild(i))
			child := node.Child(i)
			logger.Println("FieldNameForChild in in_action: ", node.FieldNameForChild(i))

			if child.Type() == "end" {
				earaseTemplate(child, previous, result)
				earaseTemplate(child.NextSibling(), previous, result)
				earaseTemplate(child.PrevSibling(), previous, result)
				break
			} else if "condition" == node.FieldNameForChild(i) {
				if_action_condition := child
				if_action_condition_content := child.Content(previous)
				logger.Println("if_action_condition_content: ", if_action_condition_content)
				earaseTemplate(if_action_condition, previous, result)
				earaseTemplate(if_action_condition.NextSibling(), previous, result)
				if if_action_condition.PrevSibling() != nil && if_action_condition.PrevSibling().Type() == "if" {
					earaseTemplate(if_action_condition.PrevSibling(), previous, result)
					earaseTemplate(if_action_condition.PrevSibling().PrevSibling(), previous, result)
				}
			} else {
				prettyPrintNode(child, previous, result)
			}
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
