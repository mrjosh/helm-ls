package yamlls

import (
	"context"
	"reflect"

	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector YamllsConnector) CallCompletion(params lsp.CompletionParams) *lsp.CompletionList {
	if yamllsConnector.Conn == nil {
		return &lsp.CompletionList{}
	}

	logger.Println("Calling yamlls for completions")
	var response = reflect.New(reflect.TypeOf(lsp.CompletionList{})).Interface()
	_, err := (*yamllsConnector.Conn).Call(context.Background(), lsp.MethodTextDocumentCompletion, params, response)
	if err != nil {
		logger.Println("Error Calling yamlls for completions", err)
	}

	logger.Println("Got completions from yamlls", response)
	return response.(*lsp.CompletionList)
}
