package cmd

import (
	"github.com/cidverse/x/pkg/app"
	"github.com/cidverse/x/pkg/common/api"
	"github.com/cidverse/x/pkg/common/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: `runs the current project`,
	Long:  `runs the current project`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "run").Msg("running command")

		// find project directory
		projectDirectory, projectDirectoryErr := filesystem.GetProjectDirectory()
		if projectDirectoryErr != nil {
			log.Fatal().Err(projectDirectoryErr).Msg(projectDirectoryErr.Error())
		}
		app.Load(projectDirectory)

		// normalize environment
		env := api.GetFullCIDEnvironment(projectDirectory)

		// actions
		app.RunStageActions("run", projectDirectory, env, args)
	},
}
