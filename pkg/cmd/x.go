package cmd

import (
	"github.com/cidverse/x/pkg/app"
	"github.com/cidverse/x/pkg/common/api"
	"github.com/cidverse/x/pkg/common/command"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	rootCmd.AddCommand(xCmd)
}

var xCmd = &cobra.Command{
	Use:   "x",
	Short: "will execute the command in the current project context.",
	Long:  `will execute the command in the current project context. Make sure you pass the args properly, ie. cid x -- go version.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "x").Str("cmd", strings.Join(args, " ")).Msg("running command")

		// find project directory and load config
		projectDir := api.FindProjectDir()
		app.Load(projectDir)

		// normalize environment
		env := api.GetCIDEnvironment(projectDir)

		// print environment
		command.RunCommand(strings.Join(args, " "), env, projectDir)
	},
}
