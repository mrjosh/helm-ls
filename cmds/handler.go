package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/sourcegraph/jsonrpc2"
	"helm.sh/helm/v3/pkg/action"
)

func NewHandler(logger *logrus.Logger) jsonrpc2.Handler {
	handler := &langHandler{
		linterName: "helm-lint",
		logger:     logger,
		request:    make(chan DocumentURI),
		client:     action.NewLint(),
	}
	go handler.linter()

	return jsonrpc2.HandlerWithError(handler.handle)
}

type langHandler struct {
	logger     *logrus.Logger
	conn       *jsonrpc2.Conn
	request    chan DocumentURI
	command    []string
	rootURI    string
	client     *action.Lint
	linterName string
}

func (h *langHandler) lint(uri DocumentURI) ([]Diagnostic, error) {

	diagnostics := make([]Diagnostic, 0)

	path := uriToPath(string(uri))
	dir, _ := filepath.Split(path)

	pathfile := ""

	paths := strings.Split(path, "/")
	h.logger.Println(paths)

	for i, p := range paths {
		if p == "templates" {
			dir = strings.Join(paths[0:i], "/")
			pathfile = strings.Join(paths[i:], "/")
		}
	}

	vals := make(map[string]interface{})
	result := h.client.Run([]string{dir}, vals)

	h.logger.Println("golangci-lint-langserver: result:", result)

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

func between(value string, a string, b string) string {
	// Get substring between two strings.
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

func after(value string, a string) string {
	// Get substring after a string.
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:]
}

func getFilePathFromLinterErr(err error) string {
	filename := ""
	fileLine := between(err.Error(), "(", ")")
	fileLineArr := strings.Split(fileLine, ":")
	file := fileLineArr[0]
	paths := strings.Split(file, "/")
	for i, p := range paths {
		if p == "templates" {
			filename = strings.Join(paths[i:], "/")
		}
	}
	return filename
}

func getDiagnosticFromLinterErr(err error) (*Diagnostic, string, error) {

	msgStr := after(err.Error(), "):")
	msg := strings.TrimSpace(msgStr)

	fileLine := between(err.Error(), "(", ")")
	fileLineArr := strings.Split(fileLine, ":")
	filename := getFilePathFromLinterErr(err)
	lineStr := fileLineArr[1]

	line, err := strconv.Atoi(lineStr)
	if err != nil {
		return nil, filename, err
	}

	return &Diagnostic{
		Range: Range{
			Start: Position{Line: line - 1},
			End:   Position{Line: line - 1},
		},
		Severity: DSError,
		Message:  msg,
	}, filename, nil
}

func (h *langHandler) linter() {

	for {

		uri, ok := <-h.request
		h.logger.Println(uri, ok)
		if !ok {
			break
		}

		diagnostics, err := h.lint(uri)
		if err != nil {
			h.logger.Printf("%s", err)

			continue
		}

		if err := h.conn.Notify(
			context.Background(),
			"textDocument/publishDiagnostics",
			&PublishDiagnosticsParams{
				URI:         uri,
				Diagnostics: diagnostics,
			}); err != nil {
			h.logger.Printf("%s", err)
		}
	}
}

func (h *langHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	h.logger.Debug("golangci-lint-langserver: request:", req)

	switch req.Method {
	case "initialize":
		return h.handleInitialize(ctx, conn, req)
	case "initialized":
		return
	case "shutdown":
		return h.handleShutdown(ctx, conn, req)
	case "textDocument/didOpen":
		return h.handleTextDocumentDidOpen(ctx, conn, req)
	case "textDocument/didClose":
		return h.handleTextDocumentDidClose(ctx, conn, req)
	case "textDocument/didChange":
		return h.handleTextDocumentDidChange(ctx, conn, req)
	case "textDocument/didSave":
		return h.handleTextDocumentDidSave(ctx, conn, req)
	case "textDocument/completion":
		return h.handleTextDocumentCompletion(ctx, conn, req)
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

func (h *langHandler) handleTextDocumentCompletion(_ context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	return nil, nil
}

func (h *langHandler) handleInitialize(_ context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	var params InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	h.rootURI = params.RootURI
	h.conn = conn
	h.command = params.InitializationOptions.Command

	return InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: TextDocumentSyncOptions{
				Change:    TDSKNone,
				OpenClose: true,
				Save:      true,
			},
		},
	}, nil
}

func (h *langHandler) handleShutdown(_ context.Context, _ *jsonrpc2.Conn, _ *jsonrpc2.Request) (result interface{}, err error) {
	close(h.request)

	return nil, nil
}

func (h *langHandler) handleTextDocumentDidOpen(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	var params DidOpenTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	h.request <- params.TextDocument.URI

	return nil, nil
}

func (h *langHandler) handleTextDocumentDidClose(_ context.Context, _ *jsonrpc2.Conn, _ *jsonrpc2.Request) (result interface{}, err error) {
	return nil, nil
}

func (h *langHandler) handleTextDocumentDidChange(_ context.Context, _ *jsonrpc2.Conn, _ *jsonrpc2.Request) (result interface{}, err error) {
	return nil, nil
}

func (h *langHandler) handleTextDocumentDidSave(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	var params DidSaveTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	h.request <- params.TextDocument.URI

	return nil, nil
}
