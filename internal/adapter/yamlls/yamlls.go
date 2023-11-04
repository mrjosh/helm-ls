package yamlls

import (
	"context"
	"os/exec"

	"github.com/mrjosh/helm-ls/internal/log"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/jsonrpc2"
)

var logger = log.GetLogger()

type YamllsConnector struct {
	Conn      *jsonrpc2.Conn
	documents *lsplocal.DocumentStore
	config    util.YamllsConfiguration
}

func NewConnector(yamllsConfiguration util.YamllsConfiguration, clientConn jsonrpc2.Conn, documents *lsplocal.DocumentStore) *YamllsConnector {
	yamllsCmd := exec.Command(yamllsConfiguration.Path, "--stdio")

	stdin, err := yamllsCmd.StdinPipe()
	if err != nil {
		logger.Println("Could not connect to stdin of yaml-language-server, some features may be missing.")
		return &YamllsConnector{}
	}
	stout, err := yamllsCmd.StdoutPipe()
	if err != nil {
		logger.Println("Could not connect to stdout of yaml-language-server, some features may be missing.")
		return &YamllsConnector{}
	}

	readWriteCloser := readWriteCloseSubprocess{
		stout,
		stdin,
	}

	err = yamllsCmd.Start()
	if err != nil {
		switch e := err.(type) {
		case *exec.Error:
			logger.Println("Could not start yaml-language-server, some features may be missing. Spawning subprocess failed.", err)
			return &YamllsConnector{}
		case *exec.ExitError:
			logger.Println("Could not start yaml-language-server, some features may be missing. Command exit rc =", e.ExitCode())
			return &YamllsConnector{}
		default:
			logger.Println("Could not start yaml-language-server, some features may be missing. Spawning subprocess failed.", err)
			return &YamllsConnector{}
		}
	}
	var yamllsConnector = YamllsConnector{}
	conn := jsonrpc2.NewConn(jsonrpc2.NewStream(readWriteCloser))
	yamllsConnector.config = yamllsConfiguration
	conn.Go(context.Background(), yamllsConnector.yamllsHandler(clientConn, documents))
	yamllsConnector.documents = documents
	yamllsConnector.Conn = &conn
	return &yamllsConnector
}
