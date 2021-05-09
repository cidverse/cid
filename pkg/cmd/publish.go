package cmd

import (
	"github.com/qubid/x/pkg/app"
	"github.com/qubid/x/pkg/common/api"
	"github.com/qubid/x/pkg/common/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(publishCmd)
	publishCmd.Flags().StringP("version", "v", "", "publish a custom version")
}

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "publishes the current project",
	Long:  `publishes the current project`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "publish").Msg("running command")

		// find project directory
		projectDirectory, projectDirectoryErr := filesystem.GetProjectDirectory()
		if projectDirectoryErr != nil {
			log.Fatal().Err(projectDirectoryErr).Msg(projectDirectoryErr.Error())
		}
		app.Load(projectDirectory)

		// normalize environment
		env := api.GetFullCIDEnvironment(projectDirectory)

		// allow to overwrite NCI_COMMIT_REF_RELEASE with a custom version
		version := cmd.Flag("version").Value.String()
		if len(version) > 0 {
			env["NCI_COMMIT_REF_RELEASE"] = version
		}

		// suggested release version
		if len(env["NCI_NEXTRELEASE_NAME"]) > 0 {
			env["NCI_COMMIT_REF_RELEASE"] = env["NCI_NEXTRELEASE_NAME"]
		}

		// actions
		log.Info().Str(`version`, env["NCI_COMMIT_REF_RELEASE"]).Msg("publishing version")
		app.RunStageActions("publish", projectDirectory, env, args)
	},
}
