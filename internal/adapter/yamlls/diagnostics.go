package yamlls

import (
	"context"
	"encoding/json"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
)

func handleDiagnostics(req jsonrpc2.Request, conn jsonrpc2.Conn, documents *lsplocal.DocumentStore) error {

	var params lsp.PublishDiagnosticsParams

	logger.Println("handleDiagnostics")
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	doc, ok := documents.Get(params.URI)
	ast := doc.Ast
	if !ok {
		// return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	filtered := filterDiagnostics(params.Diagnostics, ast)
	logger.Println(filtered)
	doc.DiagnosticsCache.Yamldiagnostics = filtered
	params.Diagnostics = doc.DiagnosticsCache.GetMergedDiagnostics()

	// logger.Println("handleDiagnostics Notify", params)
	return conn.Notify(context.Background(), lsp.MethodTextDocumentPublishDiagnostics, &params)

}

func filterDiagnostics(diagnostics []lsp.Diagnostic, ast *sitter.Tree) (filtered []lsp.Diagnostic) {

	filtered = []lsp.Diagnostic{}
	for _, diagnostic := range diagnostics {

		node := lsplocal.NodeAtPosition(ast, diagnostic.Range.Start)

		logger.Println("node Type", node.Type())
		logger.Println("diagnositc", diagnostic)
		if node.Type() == "text" {
			filtered = append(filtered, diagnostic)
		}
	}

	return filtered

}
