package cmds

import (
	"fmt"

	"github.com/mrjosh/helm-ls/internal/log"
	"github.com/spf13/cobra"
)

var logger = log.GetLogger()

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print current version of helm_ls",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(
				cmd.OutOrStdout(),
				"%s\n%s\n%s\n%s\n%s\n%s\n",
				fmt.Sprintf("Helm-ls version: %s", versionInfo.Version),
				fmt.Sprintf("Git commit: %s", versionInfo.GitCommit),
				fmt.Sprintf("Build type: %s", versionInfo.BuildType),
				fmt.Sprintf("Build time: %s", versionInfo.BuildTime),
				fmt.Sprintf("Golang: %s", versionInfo.GoVersion),
				fmt.Sprintf("Compiled by: %s", versionInfo.CompiledBy),
			)
			logger.Debug("Additional debug info will be printed")
		},
	}
}
