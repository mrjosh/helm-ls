package lsp

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mrjosh/helm-lint-ls/internal/util"
	"github.com/mrjosh/helm-lint-ls/pkg/action"
	"github.com/mrjosh/helm-lint-ls/pkg/lint/support"

	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func notifcationFromLint(ctx context.Context, conn jsonrpc2.Conn, uri uri.URI) (*jsonrpc2.Notification, error) {
	diagnostics, err := GetDiagnostics(uri)
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

func GetDiagnosticsErrors(uri uri.URI) []support.Message {

	filename := uri.Filename()
	dir, _ := filepath.Split(filename)

	paths := strings.Split(filename, "/")

	for i, p := range paths {
		if p == "templates" {
			dir = strings.Join(paths[0:i], "/")
		}
	}

	logger.Println(dir)

	client := action.NewLint()
	vals := make(map[string]interface{})

	return client.Run([]string{dir}, vals).Messages
}

func GetDiagnostics(uri uri.URI) ([]lsp.Diagnostic, error) {
	diagnostics := make([]lsp.Diagnostic, 0)

	filename := uri.Filename()
	dir, _ := filepath.Split(filename)

	pathfile := ""

	paths := strings.Split(filename, "/")

	for i, p := range paths {
		if p == "templates" {
			dir = strings.Join(paths[0:i], "/")
			pathfile = strings.Join(paths[i:], "/")
		}
	}

	logger.Println(paths)

	client := action.NewLint()
	vals := make(map[string]interface{})

	result := client.Run([]string{dir}, vals)

	logger.Println("helm lint: result:", result)

	for _, msg := range result.Messages {
		d, filename, err := GetDiagnosticFromLinterErr(msg)
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

func GetDiagnosticFromLinterErr(supMsg support.Message) (*lsp.Diagnostic, string, error) {

	var (
		err      error
		msg      string
		line     int = 1
		severity lsp.DiagnosticSeverity
		filename = getFilePathFromLinterErr(supMsg)
	)

	switch supMsg.Severity {
	case support.ErrorSev:

		severity = lsp.DiagnosticSeverityError

		fileLine := util.BetweenStrings(supMsg.Error(), "(", ")")
		fileLineArr := strings.Split(fileLine, ":")
		lineStr := fileLineArr[1]
		msgStr := util.AfterStrings(supMsg.Error(), "):")
		msg = strings.TrimSpace(msgStr)

		line, err = strconv.Atoi(lineStr)
		if err != nil {
			return nil, filename, err
		}

	case support.InfoSev:

		severity = lsp.DiagnosticSeverityInformation
		msg = supMsg.Err.Error()

	}

	return &lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{Line: uint32(line - 1)},
			End:   lsp.Position{Line: uint32(line - 1)},
		},
		Severity: severity,
		Message:  msg,
	}, filename, nil
}

func getFilePathFromLinterErr(msg support.Message) string {

	var (
		filename       string
		fileLine       = util.BetweenStrings(msg.Error(), "(", ")")
		file, _, found = strings.Cut(fileLine, ":")
	)

	if !found {
		return msg.Path
	}

	paths := strings.Split(file, "/")

	for i, p := range paths {
		if p == "templates" {
			filename = strings.Join(paths[i:], "/")
		}
	}

	return filename
}
