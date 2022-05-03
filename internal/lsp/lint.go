package lsp

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mrjosh/helm-lint-ls/internal/util"
	"helm.sh/helm/v3/pkg/action"

	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func notifcationFromLint(ctx context.Context, conn jsonrpc2.Conn, uri uri.URI) (*jsonrpc2.Notification, error) {
	diagnostics, err := getDiagnostics(uri)
	if err != nil {
		return nil, err
	}
	publishDiagnosticsParams := &lsp.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	}

	return nil, conn.Notify(
		ctx,
		lsp.MethodTextDocumentPublishDiagnostics,
		publishDiagnosticsParams,
	)
}

func getDiagnostics(uri uri.URI) ([]lsp.Diagnostic, error) {
	diagnostics := make([]lsp.Diagnostic, 0)

	path := util.URIToPath(string(uri))
	dir, _ := filepath.Split(path)

	pathfile := ""

	paths := strings.Split(path, "/")
	logger.Println(paths)

	for i, p := range paths {
		if p == "templates" {
			dir = strings.Join(paths[0:i], "/")
			pathfile = strings.Join(paths[i:], "/")
		}
	}
	client := action.NewLint()
	vals := make(map[string]interface{})
	result := client.Run([]string{dir}, vals)
	logger.Println("helm lint: result:", result)

	for _, err := range result.Errors {
		d, filename, err := getDiagnosticFromLinterErr(err)
		if err != nil {
			continue
		}
		if filename != pathfile {
			continue
		}
		diagnostics = append(diagnostics, *d)
	}

	return diagnostics, nil
}

func getDiagnosticFromLinterErr(err error) (*lsp.Diagnostic, string, error) {

	msgStr := util.AfterStrings(err.Error(), "):")
	msg := strings.TrimSpace(msgStr)

	fileLine := util.BetweenStrings(err.Error(), "(", ")")
	fileLineArr := strings.Split(fileLine, ":")
	filename := getFilePathFromLinterErr(err)
	lineStr := fileLineArr[1]

	line, err := strconv.Atoi(lineStr)
	if err != nil {
		return nil, filename, err
	}

	return &lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{Line: uint32(line - 1)},
			End:   lsp.Position{Line: uint32(line - 1)},
		},
		Severity: lsp.DiagnosticSeverityError,
		Message:  msg,
	}, filename, nil
}

func getFilePathFromLinterErr(err error) string {
	var filename string
	fileLine := util.BetweenStrings(err.Error(), "(", ")")
	file, _, found := strings.Cut(fileLine, ":")
	if !found {
		return ""
	}
	paths := strings.Split(file, "/")
	for i, p := range paths {
		if p == "templates" {
			filename = strings.Join(paths[i:], "/")
		}
	}
	return filename
}
