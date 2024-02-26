package yamlls

import (
	"context"
	"encoding/json"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
)

func (yamllsConnector *Connector) handleDiagnostics(req jsonrpc2.Request, clientConn jsonrpc2.Conn, documents *lsplocal.DocumentStore) {
	var params lsp.PublishDiagnosticsParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		logger.Println("Error handling diagnostic", err)
	}

	doc, ok := documents.Get(params.URI)
	if !ok {
		logger.Println("Error handling diagnostic. Could not get document: " + params.URI.Filename())
	}

	doc.DiagnosticsCache.SetYamlDiagnostics(filterDiagnostics(params.Diagnostics, doc.Ast.Copy(), doc.Content))
	if doc.DiagnosticsCache.ShouldShowDiagnosticsOnNewYamlDiagnostics() {
		logger.Debug("Publishing yamlls diagnostics")
		params.Diagnostics = doc.DiagnosticsCache.GetMergedDiagnostics()
		err := clientConn.Notify(context.Background(), lsp.MethodTextDocumentPublishDiagnostics, &params)
		if err != nil {
			logger.Println("Error calling yamlls for diagnostics", err)
		}
	}
}

func filterDiagnostics(diagnostics []lsp.Diagnostic, ast *sitter.Tree, content string) (filtered []lsp.Diagnostic) {
	filtered = []lsp.Diagnostic{}
	for _, diagnostic := range diagnostics {
		node := lsplocal.NodeAtPosition(ast, diagnostic.Range.Start)
		childNode := lsplocal.FindRelevantChildNode(ast.RootNode(), lsplocal.GetSitterPointForLspPos(diagnostic.Range.Start))
		if node.Type() == "text" && childNode.Type() == "text" {
			logger.Debug("Diagnostic", diagnostic)
			logger.Debug("Node", node.Content([]byte(content)))
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
	case "All mapping items must start at the same column",
		"Implicit map keys need to be followed by map values",
		"Implicit keys need to be on a single line",
		"A block sequence may not be used as an implicit map key":
		// TODO: could add a check if is is caused by includes
		return false
	case "Block scalars with more-indented leading empty lines must use an explicit indentation indicator":
		return false
		// TODO: check if this is a false positive, probably requires parsing the yaml with tree-sitter injections
		// smtp-password: |
		//   {{- if not .Values.existingSecret }}
		//   test: dsa
		//   {{- end }}

	default:
		return true
	}
}
