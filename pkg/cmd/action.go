package cmd

import (
	"github.com/cidverse/x/pkg/app"
	"github.com/cidverse/x/pkg/common/api"
	"github.com/cidverse/x/pkg/common/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(actionCmd)
}

var actionCmd = &cobra.Command{
	Use:   "action",
	Short: "runs the actions specified in the arguments",
	Long:  `runs the actions specified in the arguments`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "action").Msg("running command")

		// find project directory
		projectDirectory, projectDirectoryErr := filesystem.GetProjectDirectory()
		if projectDirectoryErr != nil {
			log.Fatal().Err(projectDirectoryErr).Msg(projectDirectoryErr.Error())
		}
		app.Load(projectDirectory)

		// normalize environment
		env := api.GetFullCIDEnvironment(projectDirectory)

		// actions
		actionName := args[0]
		action := app.FindActionByName(actionName, projectDirectory)
		if action == nil {
			log.Fatal().Str("projectDirectory", projectDirectory).Msg("can't detect the project type")
		}
		action.Execute(projectDirectory, env, args)
	},
}
