package yamlls

import (
	"context"
	"os/exec"

	"github.com/mrjosh/helm-ls/internal/log"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

var logger = log.GetLogger()

type Connector struct {
	Conn      *jsonrpc2.Conn
	config    util.YamllsConfiguration
	server    protocol.Server
	documents *lsplocal.DocumentStore
	client    protocol.Client
}

func NewConnector(yamllsConfiguration util.YamllsConfiguration, client protocol.Client, documents *lsplocal.DocumentStore) *Connector {
	yamllsCmd := exec.Command("yamlls-debug.sh", "--stdio")

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
	yamllsConnector := Connector{documents: documents, config: yamllsConfiguration, client: client}

	ctx := context.Background()
	zapLogger, _ := zap.NewProduction()
	ctx, conn, server := protocol.NewClient(ctx, yamllsConnector, jsonrpc2.NewStream(readWriteCloser), zapLogger)

	// conn.Go(context.Background(), yamllsConnector.yamllsHandler(clientConn, documents))
	yamllsConnector.Conn = &conn
	yamllsConnector.server = server
	return &yamllsConnector
}
