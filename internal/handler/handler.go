package handler

import (
	"context"
	"io"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	lsp "go.lsp.dev/protocol"
	"go.uber.org/zap"

	"github.com/mrjosh/helm-ls/internal/log"
)

var logger = log.GetLogger()

type langHandler struct {
	client          protocol.Client
	connPool        jsonrpc2.Conn
	linterName      string
	documents       *lsplocal.DocumentStore
	chartStore      *charts.ChartStore
	yamllsConnector *yamlls.Connector
	helmlsConfig    util.HelmlsConfiguration
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

func newHandler(connPool jsonrpc2.Conn, client protocol.Client) *langHandler {
	documents := lsplocal.NewDocumentStore()
	handler := &langHandler{
		client:          client,
		linterName:      "helm-lint",
		connPool:        connPool,
		documents:       documents,
		helmlsConfig:    util.DefaultConfig,
		yamllsConnector: &yamlls.Connector{},
	}
	logger.Printf("helm-lint-langserver: connections opened")

	return handler
}

// CodeAction implements protocol.Server.
func (h *langHandler) CodeAction(ctx context.Context, params *lsp.CodeActionParams) (result []lsp.CodeAction, err error) {
	panic("unimplemented")
}

// CodeLens implements protocol.Server.
func (h *langHandler) CodeLens(ctx context.Context, params *lsp.CodeLensParams) (result []lsp.CodeLens, err error) {
	panic("unimplemented")
}

// CodeLensRefresh implements protocol.Server.
func (h *langHandler) CodeLensRefresh(ctx context.Context) (err error) {
	panic("unimplemented")
}

// CodeLensResolve implements protocol.Server.
func (h *langHandler) CodeLensResolve(ctx context.Context, params *lsp.CodeLens) (result *lsp.CodeLens, err error) {
	panic("unimplemented")
}

// ColorPresentation implements protocol.Server.
func (h *langHandler) ColorPresentation(ctx context.Context, params *lsp.ColorPresentationParams) (result []lsp.ColorPresentation, err error) {
	panic("unimplemented")
}

// CompletionResolve implements protocol.Server.
func (h *langHandler) CompletionResolve(ctx context.Context, params *lsp.CompletionItem) (result *lsp.CompletionItem, err error) {
	panic("unimplemented")
}

// Declaration implements protocol.Server.
func (h *langHandler) Declaration(ctx context.Context, params *lsp.DeclarationParams) (result []lsp.Location, err error) {
	panic("unimplemented")
}

// DidChangeWorkspaceFolders implements protocol.Server.
func (h *langHandler) DidChangeWorkspaceFolders(ctx context.Context, params *lsp.DidChangeWorkspaceFoldersParams) (err error) {
	panic("unimplemented")
}

// DocumentColor implements protocol.Server.
func (h *langHandler) DocumentColor(ctx context.Context, params *lsp.DocumentColorParams) (result []lsp.ColorInformation, err error) {
	panic("unimplemented")
}

// DocumentHighlight implements protocol.Server.
func (h *langHandler) DocumentHighlight(ctx context.Context, params *lsp.DocumentHighlightParams) (result []lsp.DocumentHighlight, err error) {
	panic("unimplemented")
}

// DocumentLink implements protocol.Server.
func (h *langHandler) DocumentLink(ctx context.Context, params *lsp.DocumentLinkParams) (result []lsp.DocumentLink, err error) {
	panic("unimplemented")
}

// DocumentLinkResolve implements protocol.Server.
func (h *langHandler) DocumentLinkResolve(ctx context.Context, params *lsp.DocumentLink) (result *lsp.DocumentLink, err error) {
	panic("unimplemented")
}

// DocumentSymbol implements protocol.Server.
func (h *langHandler) DocumentSymbol(ctx context.Context, params *lsp.DocumentSymbolParams) (result []interface{}, err error) {
	panic("unimplemented")
}

// ExecuteCommand implements protocol.Server.
func (h *langHandler) ExecuteCommand(ctx context.Context, params *lsp.ExecuteCommandParams) (result interface{}, err error) {
	panic("unimplemented")
}

// Exit implements protocol.Server.
func (h *langHandler) Exit(ctx context.Context) (err error) {
	panic("unimplemented")
}

// FoldingRanges implements protocol.Server.
func (h *langHandler) FoldingRanges(ctx context.Context, params *lsp.FoldingRangeParams) (result []lsp.FoldingRange, err error) {
	panic("unimplemented")
}

// Formatting implements protocol.Server.
func (h *langHandler) Formatting(ctx context.Context, params *lsp.DocumentFormattingParams) (result []lsp.TextEdit, err error) {
	panic("unimplemented")
}

// Implementation implements protocol.Server.
func (h *langHandler) Implementation(ctx context.Context, params *lsp.ImplementationParams) (result []lsp.Location, err error) {
	panic("unimplemented")
}

// IncomingCalls implements protocol.Server.
func (h *langHandler) IncomingCalls(ctx context.Context, params *lsp.CallHierarchyIncomingCallsParams) (result []lsp.CallHierarchyIncomingCall, err error) {
	panic("unimplemented")
}

// LinkedEditingRange implements protocol.Server.
func (h *langHandler) LinkedEditingRange(ctx context.Context, params *lsp.LinkedEditingRangeParams) (result *lsp.LinkedEditingRanges, err error) {
	panic("unimplemented")
}

// LogTrace implements protocol.Server.
func (h *langHandler) LogTrace(ctx context.Context, params *lsp.LogTraceParams) (err error) {
	panic("unimplemented")
}

// Moniker implements protocol.Server.
func (h *langHandler) Moniker(ctx context.Context, params *lsp.MonikerParams) (result []lsp.Moniker, err error) {
	panic("unimplemented")
}

// OnTypeFormatting implements protocol.Server.
func (h *langHandler) OnTypeFormatting(ctx context.Context, params *lsp.DocumentOnTypeFormattingParams) (result []lsp.TextEdit, err error) {
	panic("unimplemented")
}

// OutgoingCalls implements protocol.Server.
func (h *langHandler) OutgoingCalls(ctx context.Context, params *lsp.CallHierarchyOutgoingCallsParams) (result []lsp.CallHierarchyOutgoingCall, err error) {
	panic("unimplemented")
}

// PrepareCallHierarchy implements protocol.Server.
func (h *langHandler) PrepareCallHierarchy(ctx context.Context, params *lsp.CallHierarchyPrepareParams) (result []lsp.CallHierarchyItem, err error) {
	panic("unimplemented")
}

// PrepareRename implements protocol.Server.
func (h *langHandler) PrepareRename(ctx context.Context, params *lsp.PrepareRenameParams) (result *lsp.Range, err error) {
	panic("unimplemented")
}

// RangeFormatting implements protocol.Server.
func (h *langHandler) RangeFormatting(ctx context.Context, params *lsp.DocumentRangeFormattingParams) (result []lsp.TextEdit, err error) {
	panic("unimplemented")
}

// References implements protocol.Server.
func (h *langHandler) References(ctx context.Context, params *lsp.ReferenceParams) (result []lsp.Location, err error) {
	panic("unimplemented")
}

// Rename implements protocol.Server.
func (h *langHandler) Rename(ctx context.Context, params *lsp.RenameParams) (result *lsp.WorkspaceEdit, err error) {
	panic("unimplemented")
}

// Request implements protocol.Server.
func (h *langHandler) Request(ctx context.Context, method string, params interface{}) (result interface{}, err error) {
	panic("unimplemented")
}

// SemanticTokensFull implements protocol.Server.
func (h *langHandler) SemanticTokensFull(ctx context.Context, params *lsp.SemanticTokensParams) (result *lsp.SemanticTokens, err error) {
	panic("unimplemented")
}

// SemanticTokensFullDelta implements protocol.Server.
func (h *langHandler) SemanticTokensFullDelta(ctx context.Context, params *lsp.SemanticTokensDeltaParams) (result interface{}, err error) {
	panic("unimplemented")
}

// SemanticTokensRange implements protocol.Server.
func (h *langHandler) SemanticTokensRange(ctx context.Context, params *lsp.SemanticTokensRangeParams) (result *lsp.SemanticTokens, err error) {
	panic("unimplemented")
}

// SemanticTokensRefresh implements protocol.Server.
func (h *langHandler) SemanticTokensRefresh(ctx context.Context) (err error) {
	panic("unimplemented")
}

// SetTrace implements protocol.Server.
func (h *langHandler) SetTrace(ctx context.Context, params *lsp.SetTraceParams) (err error) {
	panic("unimplemented")
}

// ShowDocument implements protocol.Server.
func (h *langHandler) ShowDocument(ctx context.Context, params *lsp.ShowDocumentParams) (result *lsp.ShowDocumentResult, err error) {
	panic("unimplemented")
}

// Shutdown implements protocol.Server.
func (h *langHandler) Shutdown(ctx context.Context) (err error) {
	return h.connPool.Close()
}

// SignatureHelp implements protocol.Server.
func (h *langHandler) SignatureHelp(ctx context.Context, params *lsp.SignatureHelpParams) (result *lsp.SignatureHelp, err error) {
	panic("unimplemented")
}

// Symbols implements protocol.Server.
func (h *langHandler) Symbols(ctx context.Context, params *lsp.WorkspaceSymbolParams) (result []lsp.SymbolInformation, err error) {
	panic("unimplemented")
}

// TypeDefinition implements protocol.Server.
func (h *langHandler) TypeDefinition(ctx context.Context, params *lsp.TypeDefinitionParams) (result []lsp.Location, err error) {
	panic("unimplemented")
}

// WillCreateFiles implements protocol.Server.
func (h *langHandler) WillCreateFiles(ctx context.Context, params *lsp.CreateFilesParams) (result *lsp.WorkspaceEdit, err error) {
	panic("unimplemented")
}

// WillDeleteFiles implements protocol.Server.
func (h *langHandler) WillDeleteFiles(ctx context.Context, params *lsp.DeleteFilesParams) (result *lsp.WorkspaceEdit, err error) {
	panic("unimplemented")
}

// WillRenameFiles implements protocol.Server.
func (h *langHandler) WillRenameFiles(ctx context.Context, params *lsp.RenameFilesParams) (result *lsp.WorkspaceEdit, err error) {
	panic("unimplemented")
}

// WillSave implements protocol.Server.
func (h *langHandler) WillSave(ctx context.Context, params *lsp.WillSaveTextDocumentParams) (err error) {
	panic("unimplemented")
}

// WillSaveWaitUntil implements protocol.Server.
func (h *langHandler) WillSaveWaitUntil(ctx context.Context, params *lsp.WillSaveTextDocumentParams) (result []lsp.TextEdit, err error) {
	panic("unimplemented")
}

// WorkDoneProgressCancel implements protocol.Server.
func (h *langHandler) WorkDoneProgressCancel(ctx context.Context, params *lsp.WorkDoneProgressCancelParams) (err error) {
	panic("unimplemented")
}
