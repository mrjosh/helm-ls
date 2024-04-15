package handler

import (
	"context"
	"fmt"

	languagefeatures "github.com/mrjosh/helm-ls/internal/language_features"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	gotemplate "github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"
	lsp "go.lsp.dev/protocol"

	"github.com/mrjosh/helm-ls/internal/documentation/godocs"
	helmdocs "github.com/mrjosh/helm-ls/internal/documentation/helm"
)

var (
	emptyItems           = make([]lsp.CompletionItem, 0)
	textCompletionsItems = make([]lsp.CompletionItem, 0)
)

func init() {
	textCompletionsItems = append(textCompletionsItems, getTextCompletionItems(godocs.TextSnippets)...)
}

func (h *langHandler) Completion(ctx context.Context, params *lsp.CompletionParams) (result *lsp.CompletionList, err error) {
	logger.Debug("Running completion with params", params)
	genericDocumentUseCase, err := h.NewGenericDocumentUseCase(params.TextDocumentPositionParams)
	if err != nil {
		return nil, err
	}

	var (
		currentNode   = lsplocal.NodeAtPosition(genericDocumentUseCase.Document.Ast, params.Position)
		pointToLoopUp = sitter.Point{
			Row:    params.Position.Line,
			Column: params.Position.Character,
		}
		relevantChildNode = lsplocal.FindRelevantChildNodeCompletion(currentNode, pointToLoopUp)
	)
	genericDocumentUseCase = genericDocumentUseCase.WithNode(relevantChildNode)

	usecases := []languagefeatures.CompletionUseCase{
		languagefeatures.NewTemplateContextFeature(genericDocumentUseCase),
		languagefeatures.NewFunctionCallFeature(genericDocumentUseCase),
	}

	for _, usecase := range usecases {
		if usecase.AppropriateForNode() {
			return usecase.Completion()
		}
	}

	word, isTextNode := completionAstParsing(genericDocumentUseCase.Document, params.Position)

	if isTextNode {
		result := make([]lsp.CompletionItem, 0)
		result = append(result, textCompletionsItems...)
		result = append(result, yamllsCompletions(ctx, h, params)...)
		logger.Debug("Sending completions ", result)
		return &protocol.CompletionList{IsIncomplete: false, Items: result}, err
	}

	logger.Println(fmt.Sprintf("Word found for completions is < %s >", word))
	items := []lsp.CompletionItem{}
	for _, v := range helmdocs.BuiltInObjects {
		items = append(items, lsp.CompletionItem{
			Label:         v.Name,
			InsertText:    "." + v.Name,
			Detail:        v.Detail,
			Documentation: v.Doc,
		})
	}
	return &lsp.CompletionList{IsIncomplete: false, Items: items}, err
}

func yamllsCompletions(ctx context.Context, h *langHandler, params *lsp.CompletionParams) []lsp.CompletionItem {
	response, err := h.yamllsConnector.CallCompletion(ctx, params)
	if err != nil {
		logger.Error("Error getting yamlls completions", err)
		return []lsp.CompletionItem{}
	}
	logger.Debug("Got completions from yamlls", response)
	return response.Items
}

func completionAstParsing(doc *lsplocal.Document, position lsp.Position) (string, bool) {
	var (
		currentNode   = lsplocal.NodeAtPosition(doc.Ast, position)
		pointToLoopUp = sitter.Point{
			Row:    position.Line,
			Column: position.Character,
		}
		relevantChildNode = lsplocal.FindRelevantChildNode(currentNode, pointToLoopUp)
		word              string
	)

	nodeType := relevantChildNode.Type()
	switch nodeType {
	case gotemplate.NodeTypeIdentifier:
		word = relevantChildNode.Content([]byte(doc.Content))
	case gotemplate.NodeTypeText, gotemplate.NodeTypeTemplate:
		return word, true
	}
	logger.Debug("word", word)
	return word, false
}

func getTextCompletionItems(gotemplateSnippet []godocs.GoTemplateSnippet) (result []lsp.CompletionItem) {
	for _, item := range gotemplateSnippet {
		result = append(result, textCompletionItem(item))
	}
	return result
}

func textCompletionItem(gotemplateSnippet godocs.GoTemplateSnippet) lsp.CompletionItem {
	return lsp.CompletionItem{
		Label: gotemplateSnippet.Name,
		// TextEdit: &lsp.TextEdit{
		// 	// Range:   lsp.Range{}, // TODO: range must contain the requested range
		// 	NewText: gotemplateSnippet.Snippet,
		// },
		InsertText:       gotemplateSnippet.Snippet,
		Detail:           gotemplateSnippet.Detail,
		Documentation:    gotemplateSnippet.Doc,
		Kind:             lsp.CompletionItemKindText,
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	}
}
