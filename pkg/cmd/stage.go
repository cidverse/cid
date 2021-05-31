package cmd

import (
	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stageCmd)
}

var stageCmd = &cobra.Command{
	Use:   "stage",
	Aliases: []string{"s"},
	Short: "runs the stage specified in the arguments",
	Long:  `runs the stage specified in the arguments`,
	Run: func(cmd *cobra.Command, args []string) {
		stage := args[0]
		log.Debug().Str("stage", stage).Msg("running stage")

		// find project directory and load config
		projectDir := api.FindProjectDir()
		app.Load(projectDir)

		// normalize environment
		env := api.GetCIDEnvironment(projectDir)

		// actions
		app.RunStageActions(stage, projectDir, env, args)
	},
}
