package yamlls

import (
	"context"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	sitter "github.com/smacker/go-tree-sitter"
	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector YamllsConnector) DocumentDidOpen(doc *lsplocal.Document, params lsp.DidOpenTextDocumentParams) {
	logger.Println("YamllsConnector DocumentDidOpen", params.TextDocument.URI)
	if yamllsConnector.Conn == nil {
		return
	}
	params.TextDocument.Text = trimTemplateForYamllsFromAst(doc.Ast, params.TextDocument.Text)

	logger.Println("Sending to yamlls", params.TextDocument.Text)

	yamllsConnector.Conn.Notify(context.Background(), lsp.MethodTextDocumentDidOpen, params)
}

func (yamllsConnector YamllsConnector) DocumentDidSave(params lsp.DidSaveTextDocumentParams) {
	if yamllsConnector.Conn == nil {
		return
	}

	yamllsConnector.Conn.Notify(context.Background(), lsp.MethodTextDocumentDidSave, params)
}

func (yamllsConnector YamllsConnector) DocumentDidChange(params lsp.DidChangeTextDocumentParams) {
	if yamllsConnector.Conn == nil {
		return
	}

	yamllsConnector.Conn.Notify(context.Background(), lsp.MethodTextDocumentDidChange, params)
}

func trimTemplateForYamlls(ast *sitter.Tree, text string) string {

	logger.Println("trimTemplateForYamlls with ast", ast.RootNode())
	logger.Println(ast.RootNode().Child(1).StartPoint())

	var runes = []rune(text)

	var result = make([]rune, 0)
	var isInTemplate = false
	var replaceNext = false
	var lineCount = 0
	var charCount = 0
	var earaseTemplate = false

	for i, char := range text {

		if char == '\n' {
			lineCount++
			charCount = 0
		} else {
			charCount++
		}

		if replaceNext && earaseTemplate {
			result = append(result, ' ')
			replaceNext = false
			if char == '}' {
				earaseTemplate = false
			}
			continue
		}

		if char == '{' && len(text) > i+1 && runes[i+1] == '{' {
			logger.Println("Template start", i)

			pos := lsp.Position{Line: uint32(lineCount), Character: uint32(charCount - 1)}

			var pointToLoopUp = sitter.Point{
				Row:    uint32(lineCount),
				Column: uint32(charCount - 1)}

			var relevantChildNode = lsplocal.FindDirectChildNodeByStart(ast.RootNode(), pointToLoopUp)
			var node = relevantChildNode

			if node == nil {
				logger.Println("Node is nil at Pos", pos)
			}

			logger.Println("Node type at", node.Type(), pos)
			logger.Println("Node ", node)

			if node.ChildCount() == 0 {
				logger.Println("child count was zero")
			}

			if node.Type() == "template" {
				earaseTemplate = true
			}
			switch node.Type() {
			case "block_action":
				logger.Println("Settign earaseTemplate to true")
				earaseTemplate = true
			case "if_action":
				logger.Println("Settign earaseTemplate to true")
				earaseTemplate = true
			}

			var sibling = node.NextNamedSibling()
			if sibling == nil {
				logger.Println("sibling is nil at Pos", pos)
			} else {

				logger.Println("sibling type", sibling.Type())
				logger.Println("sibling content", sibling.Content([]byte(text)))

				switch sibling.Type() {
				case "block_action":
					logger.Println("Settign earaseTemplate to true")
					earaseTemplate = true
				case "if_action":
					logger.Println("Settign earaseTemplate to true")
					earaseTemplate = true
				}
			}

			isInTemplate = true
		}
		if char == '}' && len(text) > i+1 && runes[i+1] == '}' {
			logger.Println("Template End", i)
			if earaseTemplate {
				result = append(result, ' ')
			} else {
				result = append(result, char)
			}
			isInTemplate = false
			replaceNext = true

			continue
		}

		if !isInTemplate {
			result = append(result, char)
			continue
		}
		if char == '\n' {
			result = append(result, char)
			continue
		}
		if earaseTemplate {
			result = append(result, ' ')
		} else {
			result = append(result, char)
		}

	}

	return string(result)
}
