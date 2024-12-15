package symboltable

import (
	"fmt"

	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	"github.com/mrjosh/helm-ls/internal/util"
	sitter "github.com/smacker/go-tree-sitter"
)

type IncludeDefinitionsVisitor struct {
	symbolTable *SymbolTable
	content     []byte
}

func NewIncludeDefinitionsVisitor(symbolTable *SymbolTable, content []byte) *IncludeDefinitionsVisitor {
	return &IncludeDefinitionsVisitor{
		symbolTable: symbolTable,
		content:     content,
	}
}

func (v *IncludeDefinitionsVisitor) Enter(node *sitter.Node) {
	if node.Type() == gotemplate.NodeTypeDefineAction {
		content := node.ChildByFieldName("name").Content(v.content)
		v.symbolTable.AddIncludeDefinition(util.RemoveQuotes(content), node.Range())
	}

	if node.Type() == gotemplate.NodeTypeFunctionCall {
		v.enterFunctionCall(node)
	}
}

func (v *IncludeDefinitionsVisitor) enterFunctionCall(node *sitter.Node) {
	includeName, err := ParseIncludeFunctionCall(node, v.content)
	if err != nil {
		return
	}

	v.symbolTable.AddIncludeReference(includeName, node.Range())
}

func ParseIncludeFunctionCall(node *sitter.Node, content []byte) (string, error) {
	if node.Type() != gotemplate.NodeTypeFunctionCall {
		return "", fmt.Errorf("node is not a function call")
	}
	functionName := node.ChildByFieldName("function").Content(content)
	if functionName != "include" && functionName != "template" { // template is not documented but was seen in the wild
		return "", fmt.Errorf("function name is not include or template")
	}
	arguments := node.ChildByFieldName("arguments")
	if arguments == nil || arguments.ChildCount() == 0 {
		return "", fmt.Errorf("no arguments")
	}
	firstArgument := arguments.Child(0)
	if firstArgument.Type() != gotemplate.NodeTypeInterpretedStringLiteral {
		return "", fmt.Errorf("first argument is not an interpreted string literal")
	}
	return util.RemoveQuotes(firstArgument.Content(content)), nil
}

func (v *IncludeDefinitionsVisitor) Exit(_ *sitter.Node)                        {}
func (v *IncludeDefinitionsVisitor) EnterContextShift(_ *sitter.Node, _ string) {}
func (v *IncludeDefinitionsVisitor) ExitContextShift(_ *sitter.Node)            {}
