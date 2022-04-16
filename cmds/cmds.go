package cmds

import (
	"github.com/spf13/cobra"
)

type VersionInfo struct {
	BranchName string
	Version    string
	GoVersion  string
	CompiledBy string
	BuildTime  string
	BuildType  string
}

var versionInfo *VersionInfo

func RegisterAndRun(vi *VersionInfo, rootCmd *cobra.Command) error {
	vi.BuildType = "Release"
	if vi.BranchName == "develop" {
		vi.BuildType = "Nightly"
	}
	versionInfo = vi
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newLintCmd())
	rootCmd.AddCommand(newTestCmd())
	return rootCmd.Execute()
}
