package cmd

import (
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/cidverse/cid/pkg/common/workflowrun"
	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/cidverse/cidverseutils/core/clioutputwriter"
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
			format, _ := cmd.Flags().GetString("format")

			// app context
			cid, err := context.NewAppContext()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to prepare app context")
				os.Exit(1)
			}

			// data
			data := clioutputwriter.TabularData{
				Headers: []string{"REPOSITORY", "ACTION", "TYPE", "SCOPE", "RULES", "DESCRIPTION"},
				Rows:    [][]interface{}{},
			}
			for _, action := range cid.Config.Registry.Actions {
				ruleEvaluation := "?/" + strconv.Itoa(len(action.Rules))
				if action.Scope == catalog.ActionScopeProject {
					ruleEvaluation = rules.EvaluateRulesAsText(action.Rules, rules.GetRuleContext(cid.Env))
				}

				data.Rows = append(data.Rows, []interface{}{
					action.Repository,
					action.Name,
					string(action.Type),
					string(action.Scope),
					ruleEvaluation,
					strings.Replace(action.Description, "\n", "", -1),
				})
			}

			// print
			writer := redact.NewProtectedWriter(nil, os.Stdout, &sync.Mutex{}, nil)
			err = clioutputwriter.PrintData(writer, data, clioutputwriter.Format(format))
			if err != nil {
				log.Fatal().Err(err).Msg("failed to print data")
				os.Exit(1)
			}
		},
	}
	cmd.Flags().StringP("format", "f", "table", "output format (table, json, csv)")

	return cmd
}

func actionRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "runs the actions specified in the arguments",
		Run: func(cmd *cobra.Command, args []string) {
			modules, _ := cmd.Flags().GetStringArray("module")

			// app context
			cid, err := context.NewAppContext()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to prepare app context")
				os.Exit(1)
			}

			// actions
			actionName := args[0]

			// pass action
			action := cid.Config.Registry.FindAction(actionName)
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
			workflowrun.RunWorkflowAction(cid.Config, &act, cid.Env, cid.ProjectDir, modules)
		},
	}
	cmd.Flags().StringArrayP("module", "m", []string{}, "limit execution to the specified module(s)")

	return cmd
}
