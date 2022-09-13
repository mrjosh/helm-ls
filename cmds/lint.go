package cmds

import (
	"fmt"
	"os"

	locallsp "github.com/mrjosh/helm-lint-ls/internal/lsp"
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

			msgs := locallsp.GetDiagnosticsErrors(uri.New(args[0] + "/templates"))
			for _, msg := range msgs {
				fmt.Println(msg.Err)
			}

			return nil
		},
	}
}
