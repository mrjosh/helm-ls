package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mrjosh/helm-lint-ls/cmds"
	"github.com/spf13/cobra"
)

var (
	BranchName string
	Version    string
	CompiledBy string
	BuildTime  string
)

func main() {

	rootCmd := &cobra.Command{
		Use: "helm_lint_ls",
		Long: `
   / / / /__  / /___ ___     / /   (_)___  / /_   / /   _____
  / /_/ / _ \/ / __  __ \   / /   / / __ \/ __/  / /   / ___/
 / __  /  __/ / / / / / /  / /___/ / / / / /_   / /___(__  ) 
/_/ /_/\___/_/_/ /_/ /_/  /_____/_/_/ /_/\__/  /_____/____/`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.SetArgs(os.Args[1:])

	vi := &cmds.VersionInfo{
		Version:    Version,
		BranchName: BranchName,
		CompiledBy: CompiledBy,
		GoVersion:  runtime.Version(),
		BuildTime:  BuildTime,
	}

	if err := cmds.RegisterAndRun(vi, rootCmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
