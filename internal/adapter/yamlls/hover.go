package yamlls

import (
	"context"
	"reflect"

	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

// Calls the Completion method of yamlls to get a fitting hover response
// TODO: clarify why the hover method of yamlls can't be used
func (yamllsConnector YamllsConnector) CallHover(params lsp.HoverParams, word string) lsp.Hover {
	if yamllsConnector.Conn == nil {
		return lsp.Hover{}
	}

	var (
		documentation    string
		response         = reflect.New(reflect.TypeOf(lsp.CompletionList{})).Interface()
		completionParams = lsp.CompletionParams{
			TextDocumentPositionParams: params.TextDocumentPositionParams,
		}
	)

	_, err := (*yamllsConnector.Conn).Call(context.Background(), lsp.MethodTextDocumentCompletion, completionParams, response)
	if err != nil {
		return util.BuildHoverResponse(documentation, lsp.Range{})
	}

	for _, completionItem := range response.(*lsp.CompletionList).Items {
		if completionItem.InsertText == word {
			documentation = completionItem.Documentation.(string)
			break
		}
	}

	return util.BuildHoverResponse(documentation, lsp.Range{})
}
