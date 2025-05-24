package yamlls

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/gobwas/glob"
	"github.com/mrjosh/helm-ls/internal/log"
	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
	"go.uber.org/zap"
)

var logger = log.GetLogger()

type Connector struct {
	config                    util.YamllsConfiguration
	server                    lsp.Server
	conn                      jsonrpc2.Conn
	documents                 *document.DocumentStore
	client                    lsp.Client
	customHandler             *CustomHandler
	EnabledForFilesGlobObject glob.Glob
}

func NewConnector(ctx context.Context,
	yamllsConfiguration util.YamllsConfiguration,
	client lsp.Client,
	documents *document.DocumentStore,
	customHandler *CustomHandler,
) *Connector {
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

	logger.Debug("Starting yaml-language-server ", yamllsConfiguration.Path)

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

	logger.Debug("Started yaml-language-server ", yamllsConfiguration.Path)

	go func() {
		io.Copy(os.Stderr, strderr)
	}()

	yamllsConnector := Connector{
		config:                    yamllsConfiguration,
		documents:                 documents,
		client:                    client,
		customHandler:             customHandler,
		EnabledForFilesGlobObject: yamllsConfiguration.GetEnabledForFilesGlobObject(),
	}

	zapLogger, _ := zap.NewProduction()
	_, conn, server := yamllsConnector.CustomNewClient(ctx, yamllsConnector, jsonrpc2.NewStream(readWriteCloser), zapLogger)

	go func() {
		<-conn.Done()
		logger.Error("yaml-language-server exited unexpectedly")
		yamllsConnector.server = nil
	}()

	yamllsConnector.server = server
	yamllsConnector.conn = conn
	return &yamllsConnector
}

func (yamllsConnector *Connector) isRelevantFile(uri lsp.URI) bool {
	doc, ok := yamllsConnector.documents.GetTemplateDoc(uri)
	if !ok {
		logger.Error("Could not find document ", uri)
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
