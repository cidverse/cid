package cmd

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"text/tabwriter"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/provenance"
	"github.com/cidverse/cidverseutils/redact"

	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
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
			// find project directory and load config
			projectDir := api.FindProjectDir()
			cfg := app.Load(projectDir)
			env := api.GetCIDEnvironment(cfg.Env, projectDir)

			// print list
			w := tabwriter.NewWriter(redact.NewProtectedWriter(nil, os.Stdout, &sync.Mutex{}, nil), 1, 1, 1, ' ', 0)
			_, _ = fmt.Fprintln(w, "WORKFLOW\tVERSION\tRULES\tSTAGES\tACTIONS")
			for _, workflow := range cfg.Registry.Workflows {
				_, _ = fmt.Fprintln(w, workflow.Name+"\t"+workflow.Version+"\t"+
					rules.EvaluateRulesAsText(workflow.Rules, rules.GetRuleContext(env))+"\t"+
					strconv.Itoa(len(workflow.Stages))+"\t"+
					strconv.Itoa(workflow.ActionCount()))
			}
			_ = w.Flush()
		},
	}

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

			// find project directory and load config
			projectDir := api.FindProjectDir()
			cfg := app.Load(projectDir)
			env := api.GetCIDEnvironment(cfg.Env, projectDir)

			if len(args) > 1 {
				// error
				_ = cmd.Help()
				os.Exit(0)
			}

			var wf *catalog.Workflow
			if len(args) == 0 {
				// evaluate rules to pick workflow
				wf = workflowrun.FirstWorkflowMatchingRules(cfg.Registry.Workflows, env)
			} else if len(args) == 1 {
				// find workflow
				wf = cfg.Registry.FindWorkflow(args[0])
			}

			if wf == nil {
				log.Fatal().Str("workflow", args[0]).Msg("workflow does not exist")
			}

			// entrypoint
			provenance.WorkflowSource = fmt.Sprintf("%s@%s", cfg.CatalogSources[wf.Repository].URI, cfg.CatalogSources[wf.Repository].SHA256)
			provenance.Workflow = fmt.Sprintf("%s@%s", wf.Name, wf.Version)

			// run
			log.Info().Str("repository", wf.Repository).Str("name", wf.Name).Str("version", wf.Version).Msg("running workflow")
			workflowrun.RunWorkflow(cfg, wf, env, projectDir, stages, modules)
		},
	}

	cmd.Flags().StringArrayP("stage", "s", []string{}, "limit execution to the specified stage(s)")
	cmd.Flags().StringArrayP("module", "m", []string{}, "limit execution to the specified module(s)")

	return cmd
}
