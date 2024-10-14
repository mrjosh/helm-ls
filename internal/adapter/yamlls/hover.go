package yamlls

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/protocol"
	lsp "go.lsp.dev/protocol"
)

// Calls the Hover method of yamlls to get a fitting hover response
// If hover returns nothing appropriate, calls yamlls for completions
//
// Yamlls can not handle hover if the schema validation returns error,
// thats why we fall back to calling completion
func (yamllsConnector Connector) CallHover(ctx context.Context, params lsp.HoverParams, word string) (*lsp.Hover, error) {
	if !yamllsConnector.shouldRun(params.TextDocumentPositionParams.TextDocument.URI) {
		return nil, nil
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

func (yamllsConnector Connector) getHoverFromCompletion(ctx context.Context, params lsp.HoverParams, word string) (response *lsp.Hover, err error) {
	var (
		resultRange      lsp.Range
		documentation    string
		completionParams = lsp.CompletionParams{
			TextDocumentPositionParams: params.TextDocumentPositionParams,
		}
	)

	word = removeTrailingColon(word)

	completionList, err := yamllsConnector.server.Completion(ctx, &completionParams)
	if err != nil {
		logger.Error("Error calling yamlls for Completion", err)
		return &lsp.Hover{}, err
	}

	for _, completionItem := range completionList.Items {
		if completionItem.InsertText == word {
			documentation = completionItem.Documentation.(string)
			resultRange = completionItem.TextEdit.Range
			break
		}
	}

	logger.Debug("Got completion for hover from yamlls", completionList)

	return protocol.BuildHoverResponse(documentation, resultRange), nil
}

func removeTrailingColon(word string) string {
	if len(word) > 2 && string(word[len(word)-1]) == ":" {
		word = word[0 : len(word)-1]
	}
	return word
}
