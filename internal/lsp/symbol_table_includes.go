package lsp

import (
	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
)

type IncludeDefinitionsVisitor struct {
	symbolTable *SymbolTable
	content     []byte
}

func (v *IncludeDefinitionsVisitor) Enter(node *sitter.Node) {
	if node.Type() == gotemplate.NodeTypeDefineAction {
		content := node.ChildByFieldName("name").Content(v.content)
		v.symbolTable.AddIncludeDefinition(util.RemoveQuotes(content), getRangeForNode(node))
	}

	// TODO: move this to separate function and use early returns
	if node.Type() == gotemplate.NodeTypeFunctionCall {
		functionName := node.ChildByFieldName("function").Content(v.content)
		if functionName == "include" {
			arguments := node.ChildByFieldName("arguments")
			if arguments.ChildCount() > 0 {
				firstArgument := arguments.Child(0)
				if firstArgument.Type() == gotemplate.NodeTypeInterpretedStringLiteral {
					content := firstArgument.Content(v.content)
					v.symbolTable.AddIncludeReference(util.RemoveQuotes(content), getRangeForNode(node))
				}
			}
		}
	}
}

func (v *IncludeDefinitionsVisitor) Exit(node *sitter.Node)                             {}
func (v *IncludeDefinitionsVisitor) EnterContextShift(node *sitter.Node, suffix string) {}
func (v *IncludeDefinitionsVisitor) ExitContextShift(node *sitter.Node)                 {}
