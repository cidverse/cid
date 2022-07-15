package cmd

import (
	"fmt"
	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/protectoutput"
	"github.com/cidverse/cid/pkg/common/workflowrun"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"text/tabwriter"
)

func init() {
	rootCmd.AddCommand(actionRootCmd)
	actionRootCmd.AddCommand(actionRunCmd)
	actionRootCmd.AddCommand(actionListCmd)
	actionRunCmd.Flags().StringArrayP("module", "m", []string{}, "limit execution to the specified module(s)")
}

var actionRootCmd = &cobra.Command{
	Use:     "action",
	Aliases: []string{"a"},
	Short:   ``,
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

var actionListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "lists all actions",
	Run: func(cmd *cobra.Command, args []string) {
		// find project directory and load config
		projectDir := api.FindProjectDir()
		cfg := app.Load(projectDir)

		// environment
		env := api.GetCIDEnvironment(projectDir)

		// print list
		w := tabwriter.NewWriter(protectoutput.NewProtectedWriter(nil, os.Stdout), 1, 1, 1, ' ', 0)
		_, _ = fmt.Fprintln(w, "REPOSITORY\tACTION\tTYPE\tSCOPE\tRULES\tDESCRIPTION")
		for _, action := range cfg.Catalog.Actions {
			_, _ = fmt.Fprintln(w, action.Repository+"\t"+
				action.Name+"\t"+
				string(action.Type)+"\t"+
				string(action.Scope)+"\t"+
				rules.EvaluateRulesAsText(action.Rules, rules.GetRuleContext(env))+"\t"+
				strings.Replace(action.Description, "\n", "", -1))
		}
		_ = w.Flush()
	},
}

var actionRunCmd = &cobra.Command{
	Use:     "run",
	Aliases: []string{"r"},
	Short:   "runs the actions specified in the arguments",
	Run: func(cmd *cobra.Command, args []string) {
		modules, _ := cmd.Flags().GetStringArray("module")

		// find project directory and load config
		projectDir := api.FindProjectDir()
		cfg := app.Load(projectDir)

		// environment
		env := api.GetCIDEnvironment(projectDir)

		// actions
		actionName := args[0]

		// pass action
		action := cfg.FindAction(actionName)
		act := config.WorkflowAction{
			Id:     action.Repository + "/" + action.Name,
			Rules:  []config.WorkflowRule{},
			Config: nil,
			Module: nil,
		}
		workflowrun.RunWorkflowAction(cfg, &act, env, projectDir, modules)
	},
}
