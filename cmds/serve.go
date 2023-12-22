package cmds

import (
	"context"
	"os"

	"github.com/mrjosh/helm-ls/internal/handler"
	"github.com/spf13/cobra"
	"go.lsp.dev/jsonrpc2"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start helm lint language server",
		RunE: func(cmd *cobra.Command, args []string) error {

			conn := jsonrpc2.NewConn(jsonrpc2.NewStream(stdrwc{}))
			handler := handler.NewHandler(conn)
			handlerSrv := jsonrpc2.HandlerServer(handler)

			return handlerSrv.ServeStream(context.Background(), conn)
		},
	}

	cmd.Flags().Bool("stdio", true, "Use stdio")

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
