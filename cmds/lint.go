package cmds

import (
	"fmt"
	"os"

	"github.com/mrjosh/helm-ls/internal/charts"
	helmlint "github.com/mrjosh/helm-ls/internal/helm_lint"
	"github.com/mrjosh/helm-ls/internal/lsp/document"
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
			documentStore := document.NewDocumentStore()
			chartStore := charts.NewChartStore(rootPath, charts.NewChart, func(chart *charts.Chart) {})
			chart, err := chartStore.GetChartForURI(rootPath)
			if err != nil {
				return err
			}

			values := documentStore.GetValuesOrEmpty(chart.ValuesFiles.MainValuesFile.URI)

			msgs := helmlint.GetDiagnostics(rootPath, values, []string{})

			for _, msg := range msgs {
				fmt.Println(msg)
			}

			return nil
		},
	}
}
