package cmd

import (
	"github.com/cidverse/x/pkg/app"
	"github.com/cidverse/x/pkg/common/api"
	"github.com/cidverse/x/pkg/common/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(packageCmd)
}

var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "packages the current project",
	Long:  `packages the current project`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "package").Msg("running command")

		// find project directory
		projectDirectory, projectDirectoryErr := filesystem.GetProjectDirectory()
		if projectDirectoryErr != nil {
			log.Fatal().Err(projectDirectoryErr).Msg(projectDirectoryErr.Error())
		}
		app.Load(projectDirectory)

		// normalize environment
		env := api.GetCIDEnvironment(projectDirectory)

		// actions
		app.RunStageActions("package", projectDirectory, env, args)
	},
}
