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
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringArrayP("module", "m", []string{}, "limit execution to the specified module(s)")
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: `runs the current project`,
	Long:  `runs the current project`,
	Run: func(cmd *cobra.Command, args []string) {
		modules, _ := cmd.Flags().GetStringArray("module")
		log.Debug().Str("command", "run").Strs("modules", modules).Msg("runCmd: execute")

		// find project directory and load config
		projectDir := api.FindProjectDir()
		app.Load(projectDir)

		// normalize environment
		env := api.GetCIDEnvironment(projectDir)

		// actions
		workflow.RunStageActions("run", modules, projectDir, filesystem.GetWorkingDirectory(), env, args)
	},
}
