package cmds

import (
	"fmt"
	"os"

	"github.com/mrjosh/helm-ls/internal/charts"
	locallsp "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
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

			rootPath := uri.New(util.FileURIScheme + args[0])
			chartStore := charts.NewChartStore(rootPath, charts.NewChart)
			chart, err := chartStore.GetChartForURI(rootPath)
			if err != nil {
				return err
			}

			msgs, err := locallsp.GetDiagnostics(uri.New(args[0]), chart.ValuesFiles.MainValuesFile.Values)
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
