package cmd

import (
	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(actionCmd)
}

var actionCmd = &cobra.Command{
	Use:   "action",
	Aliases: []string{"a", "act"},
	Short: "runs the actions specified in the arguments",
	Long:  `runs the actions specified in the arguments`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "action").Msg("running command")

		// find project directory and load config
		projectDir := api.FindProjectDir()
		app.Load(projectDir)

		// normalize environment
		env := api.GetCIDEnvironment(projectDir)

		// actions
		actionName := args[0]
		action := app.FindActionByName(actionName, projectDir, env)
		if action == nil {
			log.Fatal().Str("projectDirectory", projectDir).Str("action", actionName).Msg("can't find action by name")
		}
		action.Execute(projectDir, env, args)
	},
}
