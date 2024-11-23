package cmd

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/cmdoutput"
	"github.com/cidverse/cid/pkg/core/provenance"
	"github.com/cidverse/cidverseutils/redact"

	"github.com/cidverse/cid/pkg/common/workflowrun"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func workflowRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workflow",
		Aliases: []string{"wf"},
		Short:   ``,
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(0)
		},
	}

	cmd.AddCommand(workflowListCmd())
	cmd.AddCommand(workflowRunCmd())

	return cmd
}

func workflowListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "lists all workflows",
		Run: func(cmd *cobra.Command, args []string) {
			format, _ := cmd.Flags().GetString("format")

			// app context
			cid, err := context.NewAppContext()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to prepare app context")
				os.Exit(1)
			}

			// data
			data := cmdoutput.TabularData{
				Headers: []string{"WORKFLOW", "VERSION", "RULES", "STAGES", "ACTIONS"},
				Rows:    [][]string{},
			}
			for _, workflow := range cid.Config.Registry.Workflows {
				data.Rows = append(data.Rows, []string{
					workflow.Name,
					workflow.Version,
					rules.EvaluateRulesAsText(workflow.Rules, rules.GetRuleContext(cid.Env)),
					strconv.Itoa(len(workflow.Stages)),
					strconv.Itoa(workflow.ActionCount()),
				})
			}

			// print
			writer := redact.NewProtectedWriter(nil, os.Stdout, &sync.Mutex{}, nil)
			err = cmdoutput.PrintData(writer, data, cmdoutput.Format(format))
			if err != nil {
				log.Fatal().Err(err).Msg("failed to print data")
				os.Exit(1)
			}
		},
	}
	cmd.Flags().StringP("format", "f", "table", "output format (table, json, csv)")

	return cmd
}

func workflowRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "runs the specified workflow, requires exactly one argument",
		Run: func(cmd *cobra.Command, args []string) {
			modules, _ := cmd.Flags().GetStringArray("module")
			stages, _ := cmd.Flags().GetStringArray("stage")

			// app context
			cid, err := context.NewAppContext()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to prepare app context")
				os.Exit(1)
			}

			if len(args) > 1 {
				// error
				_ = cmd.Help()
				os.Exit(0)
			}

			var wf *catalog.Workflow
			if len(args) == 0 {
				// evaluate rules to pick workflow
				wf = workflowrun.FirstWorkflowMatchingRules(cid.Config.Registry.Workflows, cid.Env)
			} else if len(args) == 1 {
				// find workflow
				wf = cid.Config.Registry.FindWorkflow(args[0])
			}

			if wf == nil {
				log.Fatal().Str("workflow", args[0]).Msg("workflow does not exist")
			}

			// entrypoint
			provenance.WorkflowSource = fmt.Sprintf("%s@%s", cid.Config.CatalogSources[wf.Repository].URI, cid.Config.CatalogSources[wf.Repository].SHA256)
			provenance.Workflow = fmt.Sprintf("%s@%s", wf.Name, wf.Version)

			// run
			log.Info().Str("repository", wf.Repository).Str("name", wf.Name).Str("version", wf.Version).Msg("running workflow")
			workflowrun.RunWorkflow(cid.Config, wf, cid.Env, cid.ProjectDir, stages, modules)
		},
	}

	cmd.Flags().StringArrayP("stage", "s", []string{}, "limit execution to the specified stage(s)")
	cmd.Flags().StringArrayP("module", "m", []string{}, "limit execution to the specified module(s)")

	return cmd
}
