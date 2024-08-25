package handler

import (
	"context"

	"github.com/mrjosh/helm-ls/internal/charts"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"

	"go.lsp.dev/jsonrpc2"
)

func (h *langHandler) NewChartWithWatchedFiles(chart *charts.Chart) {
	logger.Debug("NewChartWithWatchedFiles ", chart.RootURI)

	uris := make([]uri.URI, 0)
	for _, valuesFile := range chart.ValuesFiles.AllValuesFiles() {
		uris = append(uris, valuesFile.URI)
	}

	go h.RegisterWatchedFiles(context.Background(), h.connPool, uris)
}

func (h *langHandler) RegisterWatchedFiles(ctx context.Context, conn jsonrpc2.Conn, files []uri.URI) {
	if conn == nil {
		return
	}
	watchers := make([]lsp.FileSystemWatcher, 0)

	for _, file := range files {
		watchers = append(watchers, lsp.FileSystemWatcher{
			GlobPattern: file.Filename(),
		})
	}

	var result any
	_, err := conn.Call(ctx, "client/registerCapability", lsp.RegistrationParams{
		Registrations: []lsp.Registration{
			{
				Method: "workspace/didChangeWatchedFiles",
				RegisterOptions: lsp.DidChangeWatchedFilesRegistrationOptions{
					Watchers: watchers,
				},
			},
		},
	}, result)
	if err != nil {
		logger.Error("Error registering watched files", err)
	}
}

func (h *langHandler) DidChangeWatchedFiles(ctx context.Context, params *lsp.DidChangeWatchedFilesParams) (err error) {
	for _, change := range params.Changes {
		h.chartStore.ReloadValuesFile(change.URI)
	}

	return nil
}
