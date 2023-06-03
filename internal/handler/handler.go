package handler

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/mrjosh/helm-ls/internal/adapter/fs"
	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/pkg/chart"
	"github.com/mrjosh/helm-ls/pkg/chartutil"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	yamlv3 "gopkg.in/yaml.v3"

	"github.com/mrjosh/helm-ls/internal/log"
)

var logger = log.GetLogger()

type langHandler struct {
	connPool        jsonrpc2.Conn
	linterName      string
	documents       *lsplocal.DocumentStore
	projectFiles    ProjectFiles
	values          chartutil.Values
	chartMetadata   chart.Metadata
	valueNode       yamlv3.Node
	chartNode       yamlv3.Node
	yamllsConnector *yamlls.YamllsConnector
}

func NewHandler(connPool jsonrpc2.Conn) jsonrpc2.Handler {
	fileStorage, _ := fs.NewFileStorage("")
	documents := lsplocal.NewDocumentStore(fileStorage)
	handler := &langHandler{
		linterName:      "helm-lint",
		connPool:        connPool,
		projectFiles:    ProjectFiles{},
		values:          make(map[string]interface{}),
		valueNode:       yamlv3.Node{},
		chartNode:       yamlv3.Node{},
		documents:       documents,
		yamllsConnector: yamlls.NewYamllsConnector("", connPool, documents),
	}
	logger.Printf("helm-lint-langserver: connections opened")
	return jsonrpc2.ReplyHandler(handler.handle)
}

func (h *langHandler) handle(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	logger.Debug("helm-lint-langserver: request:", req)

	switch req.Method() {
	case lsp.MethodInitialize:
		return h.handleInitialize(ctx, reply, req)
	case lsp.MethodInitialized:
		return reply(ctx, nil, nil)
	case lsp.MethodShutdown:
		return h.handleShutdown(ctx, reply, req)
	case lsp.MethodTextDocumentDidOpen:
		return h.handleTextDocumentDidOpen(ctx, reply, req)
	case lsp.MethodTextDocumentDidClose:
		return h.handleTextDocumentDidClose(ctx, reply, req)
	case lsp.MethodTextDocumentDidChange:
		return h.handleTextDocumentDidChange(ctx, reply, req)
	case lsp.MethodTextDocumentDidSave:
		return h.handleTextDocumentDidSave(ctx, reply, req)
	case lsp.MethodTextDocumentCompletion:
		return h.handleTextDocumentCompletion(ctx, reply, req)
	case lsp.MethodTextDocumentDefinition:
		return h.handleDefinition(ctx, reply, req)
	case lsp.MethodTextDocumentHover:
		return h.handleHover(ctx, reply, req)
	}

	return jsonrpc2.MethodNotFoundHandler(ctx, reply, req)
}

func (h *langHandler) handleInitialize(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {

	var params lsp.InitializeParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	if len(params.WorkspaceFolders) == 0 {
		return errors.New("length WorkspaceFolders is 0")
	}

	workspaceURI, err := uri.Parse(params.WorkspaceFolders[0].URI)
	h.yamllsConnector.CallInitialize(params)

	h.projectFiles = NewProjectFiles(workspaceURI, "")

	vals, err := chartutil.ReadValuesFile(h.projectFiles.ValuesFile)
	if err != nil {
		logger.Println("Error loading values.yaml file", err)
	}
	h.values = vals

	chartMetadata, err := chartutil.LoadChartfile(h.projectFiles.ChartFile)
	if err != nil {
		logger.Println("Error loading Chart.yaml file", err)
	}
	h.chartMetadata = *chartMetadata
	valueNodes, err := chartutil.ReadYamlFileToNode(h.projectFiles.ValuesFile)
	if err != nil {
		logger.Println("Error loading values.yaml file", err)
	}
	h.valueNode = valueNodes

	chartNode, err := chartutil.ReadYamlFileToNode(h.projectFiles.ChartFile)
	if err != nil {
		logger.Println("Error loading Chart.yaml file", err)
	}
	h.chartNode = chartNode

	return reply(ctx, lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: lsp.TextDocumentSyncOptions{
				Change:    lsp.TextDocumentSyncKindIncremental,
				OpenClose: true,
				Save: &lsp.SaveOptions{
					IncludeText: true,
				},
			},
			CompletionProvider: &lsp.CompletionOptions{
				TriggerCharacters: []string{".", "$."},
				ResolveProvider:   false,
			},
			HoverProvider:      true,
			DefinitionProvider: true,
		},
	}, nil)
}

func (h *langHandler) handleShutdown(_ context.Context, _ jsonrpc2.Replier, _ jsonrpc2.Request) (err error) {
	return h.connPool.Close()
}

func (h *langHandler) handleTextDocumentDidOpen(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {

	var params lsp.DidOpenTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return reply(ctx, nil, err)
	}

	doc, err := h.documents.DidOpen(params)
	if err != nil {
		logger.Println(err)
		return reply(ctx, nil, err)
	}

	h.yamllsConnector.DocumentDidOpen(doc, params)

	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}
	notification, err := lsplocal.NotifcationFromLint(ctx, h.connPool, doc)
	return reply(ctx, notification, err)
}

func (h *langHandler) handleTextDocumentDidClose(ctx context.Context, reply jsonrpc2.Replier, _ jsonrpc2.Request) (err error) {
	return reply(
		ctx,
		h.connPool.Notify(
			ctx,
			lsp.MethodTextDocumentDidClose,
			nil,
		),
		nil,
	)
}

func (h *langHandler) handleTextDocumentDidSave(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) (err error) {
	var params lsp.DidSaveTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return err
	}

	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return errors.New("Could not get document: " + params.TextDocument.URI.Filename())
	}

	h.yamllsConnector.DocumentDidSave(params)
	notification, err := lsplocal.NotifcationFromLint(ctx, h.connPool, doc)
	return reply(ctx, notification, err)
}
