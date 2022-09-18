package cmds

import (
	"github.com/mrjosh/helm-ls/internal/version"
	"github.com/spf13/cobra"
)

var versionInfo *version.BuildInfo

func Start(vi *version.BuildInfo, rootCmd *cobra.Command) error {
	vi.BuildType = "Release"
	if vi.Branch == "develop" {
		vi.BuildType = "Nightly"
	}
	versionInfo = vi
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newLintCmd())
	return rootCmd.Execute()
}
