package cmd

import (
	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/workflow"
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
		log.Info().Str(`version`, env["NCI_COMMIT_REF_RELEASE"]).Msg("publishing version")
		workflow.RunStageActions("publish", projectDir, env, args)
	},
}
