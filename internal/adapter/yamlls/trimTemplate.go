package yamlls

import (
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
)

func trimTemplateForYamllsFromAst(ast *sitter.Tree, text string) string {
	result := []byte(text)
	prettyPrintNode(ast.RootNode(), []byte(text), result)
	return string(result)
}

func prettyPrintNode(node *sitter.Node, previous []byte, result []byte) {
	childCount := node.ChildCount()

	switch node.Type() {
	case gotemplate.NodeTypeIfAction:
		trimIfAction(node, previous, result)
	case gotemplate.NodeTypeBlockAction, gotemplate.NodeTypeWithAction, gotemplate.NodeTypeRangeAction:
		trimAction(childCount, node, previous, result)
	case gotemplate.NodeTypeDefineAction:
		earaseTemplate(node, previous, result)
	case gotemplate.NodeTypeFunctionCall:
		trimFunctionCall(node, previous, result)
	case gotemplate.NodeTypeComment, gotemplate.NodeTypeVariableDefinition, gotemplate.NodeTypeAssignment:
		earaseTemplateAndSiblings(node, previous, result)
	default:
		for i := 0; i < int(childCount); i++ {
			prettyPrintNode(node.Child(i), previous, result)
		}
	}
}

func trimAction(childCount uint32, node *sitter.Node, previous []byte, result []byte) {
	for i := 0; i < int(childCount); i++ {
		child := node.Child(i)
		switch child.Type() {
		case
			gotemplate.NodeTypeAssignment,
			gotemplate.NodeTypeIf,
			gotemplate.NodeTypeSelectorExpression,
			gotemplate.NodeTypeElse,
			gotemplate.NodeTypeRange,
			gotemplate.NodeTypeFunctionCall,
			gotemplate.NodeTypeWith,
			gotemplate.NodeTypeDefine,
			gotemplate.NodeTypeOpenBraces,
			gotemplate.NodeTypeOpenBracesDash,
			gotemplate.NodeTypeCloseBraces,
			gotemplate.NodeTypeCloseBracesDash,
			gotemplate.NodeTypeEnd,
			gotemplate.NodeTypeInterpretedStringLiteral,
			gotemplate.NodeTypeBlock,
			gotemplate.NodeTypeVariableDefinition,
			gotemplate.NodeTypeVariable,
			gotemplate.NodeTypeRangeVariableDefinition:
			earaseTemplate(child, previous, result)
		default:
			prettyPrintNode(child, previous, result)
		}
	}
}

func trimIfAction(node *sitter.Node, previous []byte, result []byte) {
	curser := sitter.NewTreeCursor(node)
	curser.GoToFirstChild()
	for curser.GoToNextSibling() {
		if curser.CurrentFieldName() == gotemplate.FieldNameCondition {
			earaseTemplate(curser.CurrentNode(), previous, result)
			earaseTemplate(curser.CurrentNode().NextSibling(), previous, result)
			continue
		}
		switch curser.CurrentNode().Type() {
		case gotemplate.NodeTypeIf, gotemplate.NodeTypeElseIf:
			earaseTemplate(curser.CurrentNode(), previous, result)
			earaseTemplate(curser.CurrentNode().PrevSibling(), previous, result)
		case gotemplate.NodeTypeEnd, gotemplate.NodeTypeElse:
			earaseTemplateAndSiblings(curser.CurrentNode(), previous, result)
		default:
			prettyPrintNode(curser.CurrentNode(), previous, result)
		}
	}
	curser.Close()
}

func trimFunctionCall(node *sitter.Node, previous []byte, result []byte) {
	functionName := node.ChildByFieldName(gotemplate.FieldNameFunction)
	if functionName.Content(previous) == "include" {
		parent := node.Parent()
		if parent != nil && parent.Type() == gotemplate.NodeTypeChainedPipeline {
			earaseTemplateAndSiblings(parent, previous, result)
		}
	}
}

func earaseTemplateAndSiblings(node *sitter.Node, previous []byte, result []byte) {
	earaseTemplate(node, previous, result)
	prevSibling, nextSibling := node.PrevSibling(), node.NextSibling()
	if prevSibling != nil {
		earaseTemplate(prevSibling, previous, result)
	}
	if nextSibling != nil {
		earaseTemplate(nextSibling, previous, result)
	}
}

func earaseTemplate(node *sitter.Node, previous []byte, result []byte) {
	if node == nil {
		return
	}
	logger.Debug("Content that is erased", node.Content(previous))
	for i := range []byte(node.Content(previous)) {
		index := int(node.StartByte()) + i
		if result[index] != '\n' && result[index] != '\r' {
			result[index] = byte(' ')
		}
	}
}
