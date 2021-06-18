package cmd

import (
	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/workflow"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stageCmd)
	stageCmd.Flags().StringP("version", "v", "", "specified a custom project version")
}

var stageCmd = &cobra.Command{
	Use:     "stage",
	Aliases: []string{"s"},
	Short:   "runs the stage specified in the arguments",
	Long:    `runs the stage specified in the arguments`,
	Run: func(cmd *cobra.Command, args []string) {
		stage := args[0]
		log.Debug().Str("stage", stage).Msg("running stage")

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
		workflow.RunStageActions(stage, projectDir, filesystem.GetWorkingDirectory(), env, args)
	},
}
