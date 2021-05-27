package cmd

import (
	"github.com/cidverse/x/pkg/app"
	"github.com/cidverse/x/pkg/common/api"
	"github.com/cidverse/x/pkg/common/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringP("version", "v", "", "build a custom version")
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "builds the current project",
	Long:  `builds the current project`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "build").Msg("running command")

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
		log.Info().Str(`version`, env["NCI_COMMIT_REF_RELEASE"]).Msg("building version")
		app.RunStageActions("build", projectDirectory, env, args)
	},
}
