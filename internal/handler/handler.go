package handler

import (
	"context"
	"io"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	templatehandler "github.com/mrjosh/helm-ls/internal/handler/template_handler"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	lsp "go.lsp.dev/protocol"
	"go.uber.org/zap"

	"github.com/mrjosh/helm-ls/internal/log"
)

var logger = log.GetLogger()

type ServerHandler struct {
	client          protocol.Client
	connPool        jsonrpc2.Conn
	linterName      string
	documents       *lsplocal.DocumentStore
	chartStore      *charts.ChartStore
	yamllsConnector *yamlls.Connector
	helmlsConfig    util.HelmlsConfiguration
	langHandlers    map[lsplocal.DocumentType]LangHandler
}

func StartHandler(stream io.ReadWriteCloser) {
	logger, _ := zap.NewProduction()

	server := newHandler(nil, nil)
	_, conn, client := protocol.NewServer(context.Background(),
		server,
		jsonrpc2.NewStream(stream),
		logger,
	)
	server.connPool = conn
	server.client = client

	<-conn.Done()
}

func newHandler(connPool jsonrpc2.Conn, client protocol.Client) *ServerHandler {
	documents := lsplocal.NewDocumentStore()
	initalYamllsConnector := &yamlls.Connector{}
	handler := &ServerHandler{
		client:          client,
		linterName:      "helm-lint",
		connPool:        connPool,
		documents:       documents,
		helmlsConfig:    util.DefaultConfig,
		yamllsConnector: initalYamllsConnector,
		langHandlers: map[lsplocal.DocumentType]LangHandler{
			lsplocal.TemplateDocumentType: templatehandler.NewTemplateHandler(documents, nil, initalYamllsConnector),
		},
	}
	logger.Printf("helm-lint-langserver: connections opened")

	return handler
}

func (h *ServerHandler) setChartStrore(chartStore *charts.ChartStore) {
	h.chartStore = chartStore

	for _, handler := range h.langHandlers {
		handler.SetChartStore(chartStore)
	}
}

func (h *ServerHandler) setYamllsConnector(yamllsConnector *yamlls.Connector) {
	h.yamllsConnector = yamllsConnector

	for _, handler := range h.langHandlers {
		handler.SetYamllsConnector(yamllsConnector)
	}
}

// CodeAction implements protocol.Server.
func (h *ServerHandler) CodeAction(ctx context.Context, params *lsp.CodeActionParams) (result []lsp.CodeAction, err error) {
	logger.Error("Code action unimplemented")
	return nil, nil
}

// CodeLens implements protocol.Server.
func (h *ServerHandler) CodeLens(ctx context.Context, params *lsp.CodeLensParams) (result []lsp.CodeLens, err error) {
	logger.Error("Code lens unimplemented")
	return nil, nil
}

// CodeLensRefresh implements protocol.Server.
func (h *ServerHandler) CodeLensRefresh(ctx context.Context) (err error) {
	logger.Error("Code lens refresh unimplemented")
	return nil
}

// CodeLensResolve implements protocol.Server.
func (h *ServerHandler) CodeLensResolve(ctx context.Context, params *lsp.CodeLens) (result *lsp.CodeLens, err error) {
	logger.Error("Code lens resolve unimplemented")
	return nil, nil
}

// ColorPresentation implements protocol.Server.
func (h *ServerHandler) ColorPresentation(ctx context.Context, params *lsp.ColorPresentationParams) (result []lsp.ColorPresentation, err error) {
	logger.Error("Color presentation unimplemented")
	return nil, nil
}

// CompletionResolve implements protocol.Server.
func (h *ServerHandler) CompletionResolve(ctx context.Context, params *lsp.CompletionItem) (result *lsp.CompletionItem, err error) {
	logger.Error("Completion resolve unimplemented")
	return nil, nil
}

// Declaration implements protocol.Server.
func (h *ServerHandler) Declaration(ctx context.Context, params *lsp.DeclarationParams) (result []lsp.Location, err error) {
	logger.Error("Declaration unimplemented")
	return nil, nil
}

// DidChangeWorkspaceFolders implements protocol.Server.
func (h *ServerHandler) DidChangeWorkspaceFolders(ctx context.Context, params *lsp.DidChangeWorkspaceFoldersParams) (err error) {
	logger.Error("DidChangeWorkspaceFolders unimplemented")
	return nil
}

// DocumentColor implements protocol.Server.
func (h *ServerHandler) DocumentColor(ctx context.Context, params *lsp.DocumentColorParams) (result []lsp.ColorInformation, err error) {
	logger.Error("Document color unimplemented")
	return nil, nil
}

// DocumentHighlight implements protocol.Server.
func (h *ServerHandler) DocumentHighlight(ctx context.Context, params *lsp.DocumentHighlightParams) (result []lsp.DocumentHighlight, err error) {
	logger.Error("Document highlight unimplemented")
	return nil, nil
}

// DocumentLink implements protocol.Server.
func (h *ServerHandler) DocumentLink(ctx context.Context, params *lsp.DocumentLinkParams) (result []lsp.DocumentLink, err error) {
	logger.Error("Document link unimplemented")
	return nil, nil
}

// DocumentLinkResolve implements protocol.Server.
func (h *ServerHandler) DocumentLinkResolve(ctx context.Context, params *lsp.DocumentLink) (result *lsp.DocumentLink, err error) {
	logger.Error("Document link resolve unimplemented")
	return nil, nil
}

// ExecuteCommand implements protocol.Server.
func (h *ServerHandler) ExecuteCommand(ctx context.Context, params *lsp.ExecuteCommandParams) (result interface{}, err error) {
	logger.Error("Execute command unimplemented")
	return nil, nil
}

// Exit implements protocol.Server.
func (h *ServerHandler) Exit(ctx context.Context) (err error) {
	return nil
}

// FoldingRanges implements protocol.Server.
func (h *ServerHandler) FoldingRanges(ctx context.Context, params *lsp.FoldingRangeParams) (result []lsp.FoldingRange, err error) {
	logger.Error("Folding ranges unimplemented")
	return nil, nil
}

// Formatting implements protocol.Server.
func (h *ServerHandler) Formatting(ctx context.Context, params *lsp.DocumentFormattingParams) (result []lsp.TextEdit, err error) {
	logger.Error("Formatting unimplemented")
	return nil, nil
}

// Implementation implements protocol.Server.
func (h *ServerHandler) Implementation(ctx context.Context, params *lsp.ImplementationParams) (result []lsp.Location, err error) {
	logger.Error("Implementation unimplemented")
	return nil, nil
}

// IncomingCalls implements protocol.Server.
func (h *ServerHandler) IncomingCalls(ctx context.Context, params *lsp.CallHierarchyIncomingCallsParams) (result []lsp.CallHierarchyIncomingCall, err error) {
	logger.Error("Incoming calls unimplemented")
	return nil, nil
}

// LinkedEditingRange implements protocol.Server.
func (h *ServerHandler) LinkedEditingRange(ctx context.Context, params *lsp.LinkedEditingRangeParams) (result *lsp.LinkedEditingRanges, err error) {
	logger.Error("Linked editing range unimplemented")
	return nil, nil
}

// LogTrace implements protocol.Server.
func (h *ServerHandler) LogTrace(ctx context.Context, params *lsp.LogTraceParams) (err error) {
	logger.Error("Log trace unimplemented")
	return nil
}

// Moniker implements protocol.Server.
func (h *ServerHandler) Moniker(ctx context.Context, params *lsp.MonikerParams) (result []lsp.Moniker, err error) {
	logger.Error("Moniker unimplemented")
	return nil, nil
}

// OnTypeFormatting implements protocol.Server.
func (h *ServerHandler) OnTypeFormatting(ctx context.Context, params *lsp.DocumentOnTypeFormattingParams) (result []lsp.TextEdit, err error) {
	logger.Error("On type formatting unimplemented")
	return nil, nil
}

// OutgoingCalls implements protocol.Server.
func (h *ServerHandler) OutgoingCalls(ctx context.Context, params *lsp.CallHierarchyOutgoingCallsParams) (result []lsp.CallHierarchyOutgoingCall, err error) {
	logger.Error("Outgoing calls unimplemented")
	return nil, nil
}

// PrepareCallHierarchy implements protocol.Server.
func (h *ServerHandler) PrepareCallHierarchy(ctx context.Context, params *lsp.CallHierarchyPrepareParams) (result []lsp.CallHierarchyItem, err error) {
	logger.Error("Prepare call hierarchy unimplemented")
	return nil, nil
}

// PrepareRename implements protocol.Server.
func (h *ServerHandler) PrepareRename(ctx context.Context, params *lsp.PrepareRenameParams) (result *lsp.Range, err error) {
	logger.Error("Prepare rename unimplemented")
	return nil, nil
}

// RangeFormatting implements protocol.Server.
func (h *ServerHandler) RangeFormatting(ctx context.Context, params *lsp.DocumentRangeFormattingParams) (result []lsp.TextEdit, err error) {
	logger.Error("Range formatting unimplemented")
	return nil, nil
}

// Rename implements protocol.Server.
func (h *ServerHandler) Rename(ctx context.Context, params *lsp.RenameParams) (result *lsp.WorkspaceEdit, err error) {
	logger.Error("Rename unimplemented")
	return nil, nil
}

// Request implements protocol.Server.
func (h *ServerHandler) Request(ctx context.Context, method string, params interface{}) (result interface{}, err error) {
	logger.Error("Request unimplemented")
	return nil, nil
}

// SemanticTokensFull implements protocol.Server.
func (h *ServerHandler) SemanticTokensFull(ctx context.Context, params *lsp.SemanticTokensParams) (result *lsp.SemanticTokens, err error) {
	logger.Error("Semantic tokens full unimplemented")
	return nil, nil
}

// SemanticTokensFullDelta implements protocol.Server.
func (h *ServerHandler) SemanticTokensFullDelta(ctx context.Context, params *lsp.SemanticTokensDeltaParams) (result interface{}, err error) {
	logger.Error("Semantic tokens full delta unimplemented")
	return nil, nil
}

// SemanticTokensRange implements protocol.Server.
func (h *ServerHandler) SemanticTokensRange(ctx context.Context, params *lsp.SemanticTokensRangeParams) (result *lsp.SemanticTokens, err error) {
	logger.Error("Semantic tokens range unimplemented")
	return nil, nil
}

// SemanticTokensRefresh implements protocol.Server.
func (h *ServerHandler) SemanticTokensRefresh(ctx context.Context) (err error) {
	logger.Error("Semantic tokens refresh unimplemented")
	return nil
}

// SetTrace implements protocol.Server.
func (h *ServerHandler) SetTrace(ctx context.Context, params *lsp.SetTraceParams) (err error) {
	logger.Error("Set trace unimplemented")
	return nil
}

// ShowDocument implements protocol.Server.
func (h *ServerHandler) ShowDocument(ctx context.Context, params *lsp.ShowDocumentParams) (result *lsp.ShowDocumentResult, err error) {
	logger.Error("Show document unimplemented")
	return nil, nil
}

// Shutdown implements protocol.Server.
func (h *ServerHandler) Shutdown(ctx context.Context) (err error) {
	return h.connPool.Close()
}

// SignatureHelp implements protocol.Server.
func (h *ServerHandler) SignatureHelp(ctx context.Context, params *lsp.SignatureHelpParams) (result *lsp.SignatureHelp, err error) {
	logger.Error("Signature help unimplemented")
	return nil, nil
}

// Symbols implements protocol.Server.
func (h *ServerHandler) Symbols(ctx context.Context, params *lsp.WorkspaceSymbolParams) (result []lsp.SymbolInformation, err error) {
	logger.Error("Symbols unimplemented")
	return nil, nil
}

// TypeDefinition implements protocol.Server.
func (h *ServerHandler) TypeDefinition(ctx context.Context, params *lsp.TypeDefinitionParams) (result []lsp.Location, err error) {
	logger.Error("Type definition unimplemented")
	return nil, nil
}

// WillCreateFiles implements protocol.Server.
func (h *ServerHandler) WillCreateFiles(ctx context.Context, params *lsp.CreateFilesParams) (result *lsp.WorkspaceEdit, err error) {
	logger.Error("Will create files unimplemented")
	return nil, nil
}

// WillDeleteFiles implements protocol.Server.
func (h *ServerHandler) WillDeleteFiles(ctx context.Context, params *lsp.DeleteFilesParams) (result *lsp.WorkspaceEdit, err error) {
	logger.Error("Will delete files unimplemented")
	return nil, nil
}

// WillRenameFiles implements protocol.Server.
func (h *ServerHandler) WillRenameFiles(ctx context.Context, params *lsp.RenameFilesParams) (result *lsp.WorkspaceEdit, err error) {
	logger.Error("Will rename files unimplemented")
	return nil, nil
}

// WillSave implements protocol.Server.
func (h *ServerHandler) WillSave(ctx context.Context, params *lsp.WillSaveTextDocumentParams) (err error) {
	logger.Error("Will save unimplemented")
	return nil
}

// WillSaveWaitUntil implements protocol.Server.
func (h *ServerHandler) WillSaveWaitUntil(ctx context.Context, params *lsp.WillSaveTextDocumentParams) (result []lsp.TextEdit, err error) {
	logger.Error("Will save wait until unimplemented")
	return nil, nil
}

// WorkDoneProgressCancel implements protocol.Server.
func (h *ServerHandler) WorkDoneProgressCancel(ctx context.Context, params *lsp.WorkDoneProgressCancelParams) (err error) {
	logger.Error("Work done progress cancel unimplemented")
	return nil
}
