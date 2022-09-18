package cmds

import (
	"fmt"
	"os"

	locallsp "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/spf13/cobra"
	"go.lsp.dev/uri"
)

func newLintCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lint",
		Short: "Lint a helm project",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) == 0 {
				args = append(args, os.Getenv("PWD"))
			}

			msgs, err := locallsp.GetDiagnostics(uri.New(args[0]))
			if err != nil {
				return err
			}

			for _, msg := range msgs {
				fmt.Println(msg)
			}

			return nil
		},
	}
}
