package cmd

import (
	"github.com/qubid/x/pkg/app"
	"github.com/qubid/x/pkg/common/api"
	"github.com/qubid/x/pkg/common/filesystem"
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
		env := api.GetFullCIDEnvironment(projectDirectory)

		// actions
		app.RunStageActions("package", projectDirectory, env, args)
	},
}
