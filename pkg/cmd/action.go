package cmd

import (
	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/workflow"
	"github.com/cidverse/cid/pkg/repoanalyzer"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(actionCmd)
}

var actionCmd = &cobra.Command{
	Use:     "action",
	Aliases: []string{"a", "act"},
	Short:   "runs the actions specified in the arguments",
	Long:    `runs the actions specified in the arguments`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "action").Msg("running command")

		// find project directory and load config
		projectDir := api.FindProjectDir()
		app.Load(projectDir)

		// normalize environment
		env := api.GetCIDEnvironment(projectDir)

		// actions
		actionName := args[0]
		action, actionErr := workflow.FindWorkflowAction(actionName)
		if actionErr != nil {
			log.Fatal().Str("action", actionName).Msg("action not found")
		}

		// module-scoped actions require module information
		if action.Scope == "module" && action.Module == nil {
			log.Warn().Msg("running analyzer because of missing module info in action ...")
			modules := repoanalyzer.AnalyzeProject(projectDir, projectDir)
			action.Module = modules[0]
		}

		workflow.RunAction(action, projectDir, env, args)
	},
}
