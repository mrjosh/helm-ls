package cmds

import (
	"context"
	"os"

	"github.com/mrjosh/helm-lint-ls/internal/log"
	"github.com/mrjosh/helm-lint-ls/internal/lsp"
	"github.com/spf13/cobra"
	"go.lsp.dev/jsonrpc2"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start helm lint language server",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := jsonrpc2.NewConn(jsonrpc2.NewStream(stdrwc{}))
			handler := lsp.NewHandler(conn)
			handlerSrv := jsonrpc2.HandlerServer(handler)

			logger := log.GetLogger()
			logger.Printf("serving...")
			return handlerSrv.ServeStream(context.Background(), conn)
		},
	}

	return cmd
}

type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}

	return os.Stdout.Close()
}
