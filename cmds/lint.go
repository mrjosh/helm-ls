package cmds

import (
	"fmt"
	"os"

	"github.com/mrjosh/helm-ls/internal/charts"
	helmlint "github.com/mrjosh/helm-ls/internal/helm_lint"
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

			rootPath := uri.File(args[0])
			chartStore := charts.NewChartStore(rootPath, charts.NewChart, func(chart *charts.Chart) {})
			chart, err := chartStore.GetChartForURI(rootPath)
			if err != nil {
				return err
			}

			msgs := helmlint.GetDiagnostics(rootPath, chart.ValuesFiles.MainValuesFile.Values)

			for _, msg := range msgs {
				fmt.Println(msg)
			}

			return nil
		},
	}
}
