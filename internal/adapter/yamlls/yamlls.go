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

type Connector struct {
	Conn   *jsonrpc2.Conn
	config util.YamllsConfiguration
}

func NewConnector(yamllsConfiguration util.YamllsConfiguration, clientConn jsonrpc2.Conn, documents *lsplocal.DocumentStore) *Connector {
	yamllsCmd := exec.Command(yamllsConfiguration.Path, "--stdio")

	stdin, err := yamllsCmd.StdinPipe()
	if err != nil {
		logger.Error("Could not connect to stdin of yaml-language-server, some features may be missing.")
		return &Connector{}
	}
	stout, err := yamllsCmd.StdoutPipe()
	if err != nil {
		logger.Error("Could not connect to stdout of yaml-language-server, some features may be missing.")
		return &Connector{}
	}

	readWriteCloser := readWriteCloseSubprocess{
		stout,
		stdin,
	}

	err = yamllsCmd.Start()
	if err != nil {
		switch e := err.(type) {
		case *exec.Error:
			logger.Error("Could not start yaml-language-server, some features may be missing. Spawning subprocess failed.", err)
			return &Connector{}
		case *exec.ExitError:
			logger.Error("Could not start yaml-language-server, some features may be missing. Command exit rc =", e.ExitCode())
			return &Connector{}
		default:
			logger.Error("Could not start yaml-language-server, some features may be missing. Spawning subprocess failed.", err)
			return &Connector{}
		}
	}
	var yamllsConnector = Connector{}
	conn := jsonrpc2.NewConn(jsonrpc2.NewStream(readWriteCloser))
	yamllsConnector.config = yamllsConfiguration
	conn.Go(context.Background(), yamllsConnector.yamllsHandler(clientConn, documents))
	yamllsConnector.Conn = &conn
	return &yamllsConnector
}
