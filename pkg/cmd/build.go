package cmd

import (
	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
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

		// find project directory and load config
		projectDir := api.FindProjectDir()
		app.Load(projectDir)

		// normalize environment
		env := api.GetCIDEnvironment(projectDir)

		// allow to overwrite NCI_COMMIT_REF_RELEASE with a custom version
		version := cmd.Flag("version").Value.String()
		if len(version) > 0 {
			// manually overwrite version
			env["NCI_COMMIT_REF_RELEASE"] = version
		} else if len(env["NCI_NEXTRELEASE_NAME"]) > 0 {
			// take suggested release version
			env["NCI_COMMIT_REF_RELEASE"] = env["NCI_NEXTRELEASE_NAME"]
		}

		// actions
		log.Info().Str(`version`, env["NCI_COMMIT_REF_RELEASE"]).Msg("building version")
		app.RunStageActions("build", projectDir, env, args)
	},
}
