package yamlls

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/mrjosh/helm-ls/internal/log"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	lsp "go.lsp.dev/protocol"
	"go.uber.org/zap"
)

var logger = log.GetLogger()

type Connector struct {
	config    util.YamllsConfiguration
	server    protocol.Server
	documents *lsplocal.DocumentStore
	client    protocol.Client
}

func NewConnector(ctx context.Context, yamllsConfiguration util.YamllsConfiguration, client protocol.Client, documents *lsplocal.DocumentStore) *Connector {
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

	strderr, err := yamllsCmd.StderrPipe()
	if err != nil {
		logger.Error("Could not connect to stderr of yaml-language-server, some features may be missing.")
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

	go func() {
		io.Copy(os.Stderr, strderr)
	}()

	yamllsConnector := Connector{
		config:    yamllsConfiguration,
		documents: documents,
		client:    client,
	}

	zapLogger, _ := zap.NewProduction()
	_, _, server := protocol.NewClient(ctx, yamllsConnector, jsonrpc2.NewStream(readWriteCloser), zapLogger)

	yamllsConnector.server = server
	return &yamllsConnector
}

func (yamllsConnector *Connector) isRelevantFile(uri lsp.URI) bool {
	doc, ok := yamllsConnector.documents.Get(uri)
	if !ok {
		logger.Error("Could not find document", uri)
		return true
	}
	return doc.IsYaml
}

func (yamllsConnector *Connector) shouldRun(uri lsp.DocumentURI) bool {
	if yamllsConnector.server == nil {
		return false
	}
	return yamllsConnector.isRelevantFile(uri)
}
