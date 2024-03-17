package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mrjosh/helm-ls/cmds"
	"github.com/mrjosh/helm-ls/internal/version"
	"github.com/spf13/cobra"
)

var (
	BranchName string
	Version    string
	GitCommit  string
	CompiledBy string
	BuildTime  string
)

func main() {
	rootCmd := &cobra.Command{
		Use: "helm_ls",
		Long: `
  /\  /\___| |_ __ ___   / / ___ 
 / /_/ / _ \ | '_ ' _ \ / / / __|
/ __  /  __/ | | | | | / /__\__ \
\/ /_/ \___|_|_| |_| |_\____/___/`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	rootCmd.SetArgs(os.Args[1:])

	vi := &version.BuildInfo{
		Version:    Version,
		Branch:     BranchName,
		GitCommit:  GitCommit,
		CompiledBy: CompiledBy,
		GoVersion:  runtime.Version(),
		BuildTime:  BuildTime,
	}

	if err := cmds.Start(vi, rootCmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
