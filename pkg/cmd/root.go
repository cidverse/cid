package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   `cid`,
	Short: `cid is a cli to run pipeline actions locally and as part of your ci/cd process`,
	Long:  `cid is a cli to run pipeline actions locally and as part of your ci/cd process`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
