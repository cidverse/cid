package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/workflowrun"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func actionRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "action",
		Aliases: []string{"a"},
		Short:   ``,
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(0)
		},
	}

	cmd.AddCommand(actionListCmd())
	cmd.AddCommand(actionRunCmd())

	return cmd
}

func actionListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "lists all actions",
		Run: func(cmd *cobra.Command, args []string) {
			// find project directory and load config
			projectDir := api.FindProjectDir()
			cfg := app.Load(projectDir)
			env := api.GetCIDEnvironment(cfg.Env, projectDir)

			// print list
			w := tabwriter.NewWriter(redact.NewProtectedWriter(nil, os.Stdout, &sync.Mutex{}, nil), 1, 1, 1, ' ', 0)
			_, _ = fmt.Fprintln(w, "REPOSITORY\tACTION\tTYPE\tSCOPE\tRULES\tDESCRIPTION")
			for _, action := range cfg.Registry.Actions {
				ruleEvaluation := "?/" + strconv.Itoa(len(action.Rules))
				if action.Scope == catalog.ActionScopeProject {
					ruleEvaluation = rules.EvaluateRulesAsText(action.Rules, rules.GetRuleContext(env))
				}

				_, _ = fmt.Fprintln(w, action.Repository+"\t"+
					action.Name+"\t"+
					string(action.Type)+"\t"+
					string(action.Scope)+"\t"+
					ruleEvaluation+"\t"+
					strings.Replace(action.Description, "\n", "", -1))
			}
			_ = w.Flush()
		},
	}

	return cmd
}

func actionRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "runs the actions specified in the arguments",
		Run: func(cmd *cobra.Command, args []string) {
			modules, _ := cmd.Flags().GetStringArray("module")

			// find project directory and load config
			projectDir := api.FindProjectDir()
			cfg := app.Load(projectDir)
			env := api.GetCIDEnvironment(cfg.Env, projectDir)

			// actions
			actionName := args[0]

			// pass action
			action := cfg.Registry.FindAction(actionName)
			if action == nil {
				log.Error().Str("action", actionName).Msg("action is not known")
				os.Exit(1)
			}
			act := catalog.WorkflowAction{
				ID:     action.Repository + "/" + action.Name,
				Rules:  []catalog.WorkflowRule{},
				Config: nil,
				Module: nil,
			}
			workflowrun.RunWorkflowAction(cfg, &act, env, projectDir, modules)
		},
	}

	cmd.Flags().StringArrayP("module", "m", []string{}, "limit execution to the specified module(s)")

	return cmd
}
