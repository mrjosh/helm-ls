package handler

import (
	"context"

	lsp "go.lsp.dev/protocol"
)

func (h *langHandler) publishDiagnostics(ctx context.Context, notifications []lsp.PublishDiagnosticsParams) {
	for _, notification := range notifications {
		logger.Debug("Publishing diagnostics notification ", notification)
		err := h.client.PublishDiagnostics(ctx, &notification)
		if err != nil {
			logger.Error("Error publishing diagnostics ", err)
		}
	}
}
