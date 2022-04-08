package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
	rootCmd.AddCommand(newServeCmd(os.Stdout))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
