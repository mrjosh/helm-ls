package yamlls

import (
	"context"
	"reflect"

	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

// Calls the Hover method of yamlls to get a fitting hover response
// If hover returns nothing appropriate, calls yamlls for completions
func (yamllsConnector Connector) CallHover(ctx context.Context, params lsp.HoverParams, word string) (*lsp.Hover, error) {
	if yamllsConnector.Conn == nil {
		return &lsp.Hover{}, nil
	}

	hoverResponse, err := (yamllsConnector).getHoverFromHover(ctx, params)
	if err != nil {
		return hoverResponse, err
	}

	if hoverResponse.Contents.Value != "" {
		return hoverResponse, nil
	}
	return (yamllsConnector).getHoverFromCompletion(ctx, params, word)
}

func (yamllsConnector Connector) getHoverFromHover(ctx context.Context, params lsp.HoverParams) (*lsp.Hover, error) {

	var hoverResponse = reflect.New(reflect.TypeOf(lsp.Hover{})).Interface()
	_, err := (*yamllsConnector.Conn).Call(ctx, lsp.MethodTextDocumentHover, params, hoverResponse)
	if err != nil {
		logger.Error("Error calling yamlls for hover", err)
		return &lsp.Hover{}, err
	}
	logger.Debug("Got hover from yamlls", hoverResponse.(*lsp.Hover).Contents.Value)
	return hoverResponse.(*lsp.Hover), nil
}

func (yamllsConnector Connector) getHoverFromCompletion(ctx context.Context, params lsp.HoverParams, word string) (*lsp.Hover, error) {
	var (
		err                error
		documentation      string
		completionResponse = reflect.New(reflect.TypeOf(lsp.CompletionList{})).Interface()
		completionParams   = lsp.CompletionParams{
			TextDocumentPositionParams: params.TextDocumentPositionParams,
		}
	)
	_, err = (*yamllsConnector.Conn).Call(ctx, lsp.MethodTextDocumentCompletion, completionParams, completionResponse)
	if err != nil {
		logger.Error("Error calling yamlls for Completion", err)
		return &lsp.Hover{}, err
	}

	for _, completionItem := range completionResponse.(*lsp.CompletionList).Items {
		if completionItem.InsertText == word {
			documentation = completionItem.Documentation.(string)
			break
		}
	}

	response := util.BuildHoverResponse(documentation, lsp.Range{})
	return &response, nil
}
