package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/sourcegraph/jsonrpc2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/lint/support"
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
			pathfile = pathfile[0 : len(pathfile)-1]
		}
	}

	vals := make(map[string]interface{})
	result := h.client.Run([]string{dir}, vals)

	h.logger.Println("golangci-lint-langserver: result:", result)

	for _, d := range result.Messages {

		if d.Path != "" {

			if strings.Contains(d.Error(), "parse error at") {

				d, filename, err := h.getDiagnosticFromLinterMessageWithoutLineNum(&d)
				if err != nil {
					continue
				}

				log.Println(fmt.Sprintf("[%s,%s]", filename, pathfile))

				if filename != pathfile {
					continue
				}

				diagnostics = append(diagnostics, *d)
				continue
			}

			// find lint error that has line number in them
			if strings.Contains(d.Error(), "line") {

				d, filename, err := h.getDiagnosticFromLinterMessage(&d)
				if err != nil {
					continue
				}

				log.Println(fmt.Sprintf("[%s,%s]", filename, pathfile))

				if filename != pathfile {
					continue
				}

				diagnostics = append(diagnostics, *d)
				continue
			}

		}

	}

	return diagnostics, nil
}

func getFilePathFromLinterMessage(msg *support.Message) string {
	filename := ""
	paths := strings.Split(msg.Path, "/")
	for i, p := range paths {
		if p == "templates" {
			filename = strings.Join(paths[i:], "/")
		}
	}
	return filename
}

func (h *langHandler) getDiagnosticFromLinterMessageWithoutLineNum(lintMsg *support.Message) (*Diagnostic, string, error) {

	fileLine := between(lintMsg.Error(), "(", ")")
	fileLineArr := strings.Split(fileLine, ":")
	filename := getFilePathFromLinterErr(lintMsg.Err)
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
		Message:  lintMsg.Error(),
	}, filename, nil
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

func (h *langHandler) getDiagnosticFromLinterMessage(lintMsg *support.Message) (*Diagnostic, string, error) {

	lineNumber, err := h.findLineNumber(lintMsg.Error())
	if err != nil {
		h.logger.Printf("%s", err)
		return nil, "", err
	}

	filename := getFilePathFromLinterMessage(lintMsg)

	return &Diagnostic{
		Range: Range{
			Start: Position{Line: lineNumber - 1},
			End:   Position{Line: lineNumber - 1},
		},
		Severity: DSError,
		Message:  lintMsg.Error(),
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

func (h *langHandler) findLineNumber(str string) (int, error) {

	re := regexp.MustCompile(`(?s)line \d+`)
	if len(re.FindStringIndex(str)) > 0 {
		match := re.FindString(str)
		lineNumber := strings.ReplaceAll(match, "line", "")
		return strconv.Atoi(strings.TrimSpace(lineNumber))
	}

	return 0, fmt.Errorf("could not find line number from this err: %v", str)
}

func (h *langHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	h.logger.Debug("golangci-lint-langserver: request:", req)

	switch req.Method {
	case "initialize":
		return h.handleInitialize(ctx, conn, req)
	case "initialized":
		return
	case "$/cancelRequest":
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
