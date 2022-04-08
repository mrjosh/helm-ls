package main

import (
	"context"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/cobra"
)

func newServeCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start helm lint language server",
		RunE: func(cmd *cobra.Command, args []string) error {

			logger := logrus.New()
			logger.SetFormatter(&logrus.JSONFormatter{})

			handler := NewHandler(logger)

			var connOpt []jsonrpc2.ConnOpt
			logger.Printf("helm-lint-langserver: connections opened")

			<-jsonrpc2.NewConn(
				context.Background(),
				jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{}),
				handler,
				connOpt...,
			).DisconnectNotify()

			return nil
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
