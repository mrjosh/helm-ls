package yamlls

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"go.lsp.dev/jsonrpc2"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	lsp "go.lsp.dev/protocol"

	"github.com/mrjosh/helm-ls/internal/log"
)

var logger = log.GetLogger()

type YamllsConnector struct {
	Conn      jsonrpc2.Conn
	documents *lsplocal.DocumentStore
}

func NewYamllsConnector(workingDir string, clientConn jsonrpc2.Conn, documents *lsplocal.DocumentStore) *YamllsConnector {
	// yamllsCmd := exec.Command("yaml-language-server", "--stdio")
	yamllsCmd := exec.Command("debug-lsp.sh")

	stdin, err := yamllsCmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	stout, err := yamllsCmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	// stder, err := yamllsCmd.StderrPipe()
	// if err != nil {
	// 	panic(err)
	// }
	// go readStuff(bufio.NewScanner(stder))

	readWriteCloser := readWriteCloseSubprocess{
		stout,
		stdin,
	}

	err = yamllsCmd.Start()
	if err != nil {
		switch e := err.(type) {
		case *exec.Error:
			fmt.Println("failed executing:", err)
			panic(err)
		case *exec.ExitError:
			fmt.Println("command exit rc =", e.ExitCode())
			panic(err)
		default:
			panic(err)
		}
	}
	var yamllsConnector = YamllsConnector{}
	conn := jsonrpc2.NewConn(jsonrpc2.NewStream(readWriteCloser))
	conn.Go(context.Background(), testHandler(clientConn, documents))
	yamllsConnector.documents = documents
	yamllsConnector.Conn = conn
	return &yamllsConnector
}
func testHandler(clientConn jsonrpc2.Conn, documents *lsplocal.DocumentStore) jsonrpc2.Handler {
	return func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {

		logger.Println("Handler called ", req.Method())
		// logger.Println("Handler called", req.Method())
		// logger.Println("Handler called", params)

		if req.Method() == lsp.MethodTextDocumentPublishDiagnostics {
			var params lsp.PublishDiagnosticsParams

			if err := json.Unmarshal(req.Params(), &params); err != nil {
				return err
			}
			err := handleDiagnostics(req, clientConn, documents)

			if err != nil {
				logger.Println(err)
			}
		}
		if req.Method() == lsp.MethodWorkspaceConfiguration {
			var params lsp.ConfigurationParams

			if err := json.Unmarshal(req.Params(), &params); err != nil {
				return err
			}

			settings := [5]interface{}{YamllsSettings{Schemas: map[string]string{"kubernetes": "**"}}}

			logger.Println("Handler called", params)
			return reply(ctx, settings, nil)
		}

		return reply(ctx, true, nil)
	}

}

func readStuff(scanner *bufio.Scanner) {
	logger.Println("Start scanner")
	for scanner.Scan() {
		logger.Println("scanner found")
		logger.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
