package yamlhandler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mrjosh/helm-ls/internal/lsp/symboltable"
	"github.com/mrjosh/helm-ls/internal/protocol"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
)

// Hover implements handler.LangHandler.
func (h *YamlHandler) Hover(ctx context.Context, params *lsp.HoverParams) (*lsp.Hover, error) {
	logger.Debug("YamlHandler Hover", params)

	yamlResult, yamllsErr := h.yamllsConnector.CallHover(ctx, *params)
	path, yamlPathErr := h.getYamlPath(params.TextDocument.URI, params.Position)
	templateContext := symboltable.TemplateContextFromYAMLPath(path)

	if yamlPathErr != nil {
		return yamlResult, errors.Join(yamllsErr, yamlPathErr)
	}

	valuesResult, valuesErr := h.otherValuesFilesHover(params, templateContext)

	if yamlResult == nil {
		return protocol.BuildHoverResponse(templateContext.Format()+"\n\n"+valuesResult, lsp.Range{}), errors.Join(yamllsErr, yamlPathErr, valuesErr)
	}

	yamlResult.Contents.Value = yamlResult.Contents.Value + "\n\n" + templateContext.Format() + "\n\n" + valuesResult

	// IDEA: get the definitions from other values files and include comments (documentation) in the result

	return yamlResult, errors.Join(yamllsErr, yamlPathErr, valuesErr)
}

func (h *YamlHandler) otherValuesFilesHover(params *lsp.HoverParams, templateContext symboltable.TemplateContext) (string, error) {
	chart, err := h.chartStore.GetChartForDoc(params.TextDocument.URI)
	if err != nil {
		return "", fmt.Errorf("getting Hover failed for document, could not get chart for document: %w", err)
	}
	var (
		valuesFiles  = chart.ResolveValueFiles(templateContext, h.chartStore)
		hoverResults = protocol.HoverResultsWithFiles{}
	)
	for _, valuesFiles := range valuesFiles {
		for _, valuesFile := range valuesFiles.ValuesFiles.AllValuesFiles() {
			logger.Debug(fmt.Sprintf("Looking for selector: %s in values %v", strings.Join(valuesFiles.Selector, "."), valuesFile.Values))
			result, err := util.GetTableOrValueForSelector(valuesFile.Values, valuesFiles.Selector)

			if err == nil {
				hoverResults = append(hoverResults, protocol.HoverResultWithFile{URI: valuesFile.URI, Value: result})
			}
		}
	}
	valuesResult := hoverResults.FormatYaml(h.chartStore.RootURI)
	return valuesResult, nil
}
