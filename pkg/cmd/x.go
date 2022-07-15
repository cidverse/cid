package cmd

import (
	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	rootCmd.AddCommand(xCmd)
}

var xCmd = &cobra.Command{
	Use:     "x",
	Short:   `will execute the command in the current project context.`,
	Example: `cid x -- go version`,
	Run: func(cmd *cobra.Command, args []string) {
		// find project directory and load config
		projectDir := api.FindProjectDir()
		app.Load(projectDir)
		env := api.GetCIDEnvironment(projectDir)

		// print environment
		command.RunCommand(strings.Join(args, " "), env, projectDir)
	},
}
