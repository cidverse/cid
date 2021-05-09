package cmd

import (
	"github.com/qubid/x/pkg/app"
	"github.com/qubid/x/pkg/common/api"
	"github.com/qubid/x/pkg/common/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "tests the current project",
	Long:  `tests the current project`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "test").Msg("running command")

		// find project directory
		projectDirectory, projectDirectoryErr := filesystem.GetProjectDirectory()
		if projectDirectoryErr != nil {
			log.Fatal().Err(projectDirectoryErr).Msg(projectDirectoryErr.Error())
		}
		app.Load(projectDirectory)

		// normalize environment
		env := api.GetFullCIDEnvironment(projectDirectory)

		// actions
		app.RunStageActions("test", projectDirectory, env, args)
	},
}
