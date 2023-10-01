package yamlls

import (
	"context"
	"encoding/json"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
)

func handleDiagnostics(req jsonrpc2.Request, clientConn jsonrpc2.Conn, documents *lsplocal.DocumentStore) {
	var params lsp.PublishDiagnosticsParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		logger.Println("Error handling diagnostic", err)
	}

	doc, ok := documents.Get(params.URI)
	if !ok {
		logger.Println("Error handling diagnostic. Could not get document: " + params.URI.Filename())
	}
	doc.DiagnosticsCache.Yamldiagnostics = filterDiagnostics(params.Diagnostics, doc.Ast, doc.Content)
	params.Diagnostics = doc.DiagnosticsCache.GetMergedDiagnostics()

	clientConn.Notify(context.Background(), lsp.MethodTextDocumentPublishDiagnostics, &params)
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
	logger.Println("Diagnostic", diagnostic.Message)
	switch diagnostic.Message {
	case "Map keys must be unique":
		return !lsplocal.IsInElseBranch(node)
	case "All mapping items must start at the same column", "Implicit map keys need to be followed by map values", "Implicit keys need to be on a single line", "A block sequence may not be used as an implicit map key":
		// TODO: could add a check if is is caused by includes
		return false

	default:
		return true
	}

}
