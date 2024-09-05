package yamlls

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"
	lsp "go.lsp.dev/protocol"
)

func (c Connector) PublishDiagnostics(ctx context.Context, params *protocol.PublishDiagnosticsParams) (err error) {
	doc, ok := c.documents.Get(params.URI)
	if !ok {
		logger.Println("Error handling diagnostic. Could not get document: " + params.URI.Filename())
		return fmt.Errorf("Could not get document: %s", params.URI.Filename())
	}

	doc.DiagnosticsCache.SetYamlDiagnostics(filterDiagnostics(params.Diagnostics, doc.Ast.Copy(), doc.Content))
	if doc.DiagnosticsCache.ShouldShowDiagnosticsOnNewYamlDiagnostics() {
		logger.Debug("Publishing yamlls diagnostics")
		params.Diagnostics = doc.DiagnosticsCache.GetMergedDiagnostics()
		err := c.client.PublishDiagnostics(ctx, params)
		if err != nil {
			logger.Println("Error calling yamlls for diagnostics", err)
		}
	}

	return nil
}

func filterDiagnostics(diagnostics []lsp.Diagnostic, ast *sitter.Tree, content []byte) (filtered []lsp.Diagnostic) {
	filtered = []lsp.Diagnostic{}

	for _, diagnostic := range diagnostics {
		node := lsplocal.NodeAtPosition(ast, diagnostic.Range.Start)
		childNode := lsplocal.FindRelevantChildNode(ast.RootNode(), lsplocal.GetSitterPointForLspPos(diagnostic.Range.Start))

		if node.Type() == "text" && childNode.Type() == "text" {
			logger.Debug("Diagnostic", diagnostic)
			logger.Debug("Node", node.Content(content))

			if diagnisticIsRelevant(diagnostic, childNode) {
				diagnostic.Message = "Yamlls: " + diagnostic.Message
				filtered = append(filtered, diagnostic)
			}
		}
	}

	return filtered
}

func diagnisticIsRelevant(diagnostic lsp.Diagnostic, node *sitter.Node) bool {
	logger.Debug("Checking if diagnostic is relevant", diagnostic.Message)

	switch diagnostic.Message {
	case "Map keys must be unique":
		return !lsplocal.IsInElseBranch(node)
	case "All mapping items must start at the same column":
		// unknown what exactly this is, only causes one error in bitnami/charts
		return false
	case "Implicit map keys need to be followed by map values", "A block sequence may not be used as an implicit map key", "Implicit keys need to be on a single line":
		// still breaks with
		// params:
		// {{- range $key, $value := .params }}
		// {{ $key }}:
		//   {{- range $value }}
		//   - {{ . | quote }}
		//   {{- end }}
		// {{- end }}
		return false && !lsplocal.IsInElseBranch(node)
	case "Block scalars with more-indented leading empty lines must use an explicit indentation indicator":
		// TODO: check if this is a false positive, probably requires parsing the yaml with tree-sitter injections
		// smtp-password: |
		//   {{- if not .Values.existingSecret }}
		//   test: dsa
		//   {{- end }}
		return false
	default:
		// TODO: remove this once the tree-sitter grammar behavior for windows newlines is the same as for unix
		if strings.HasPrefix(diagnostic.Message, "Incorrect type. Expected") && runtime.GOOS == "windows" {
			return false
		}

		return true
	}
}
