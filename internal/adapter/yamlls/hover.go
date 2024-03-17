package yamlls

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

// Calls the Hover method of yamlls to get a fitting hover response
// If hover returns nothing appropriate, calls yamlls for completions
func (yamllsConnector Connector) CallHover(ctx context.Context, params lsp.HoverParams, word string) (*lsp.Hover, error) {
	if yamllsConnector.server == nil {
		return &lsp.Hover{}, nil
	}

	hoverResponse, err := yamllsConnector.server.Hover(ctx, &params)
	if err != nil {
		return hoverResponse, err
	}

	if hoverResponse != nil && hoverResponse.Contents.Value != "" {
		return hoverResponse, nil
	}
	return (yamllsConnector).getHoverFromCompletion(ctx, params, word)
}

func (yamllsConnector Connector) getHoverFromCompletion(ctx context.Context, params lsp.HoverParams, word string) (*lsp.Hover, error) {
	var (
		err              error
		documentation    string
		completionParams = lsp.CompletionParams{
			TextDocumentPositionParams: params.TextDocumentPositionParams,
		}
	)

	completionList, err := yamllsConnector.server.Completion(ctx, &completionParams)
	if err != nil {
		logger.Error("Error calling yamlls for Completion", err)
		return &lsp.Hover{}, err
	}

	for _, completionItem := range completionList.Items {
		if completionItem.InsertText == word {
			documentation = completionItem.Documentation.(string)
			break
		}
	}

	response := util.BuildHoverResponse(documentation, lsp.Range{})
	return response, nil
}
