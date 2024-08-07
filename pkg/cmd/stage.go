package cmd

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"text/tabwriter"

	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/spf13/cobra"
)

func stageRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stage",
		Aliases: []string{"s"},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(0)
		},
	}

	cmd.AddCommand(stageListCmd())

	return cmd
}

func stageListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "lists all stages",
		Run: func(cmd *cobra.Command, args []string) {
			// find project directory and load config
			projectDir := api.FindProjectDir()
			cfg := app.Load(projectDir)
			env := api.GetCIDEnvironment(cfg.Env, projectDir)

			// print list
			w := tabwriter.NewWriter(redact.NewProtectedWriter(nil, os.Stdout, &sync.Mutex{}, nil), 1, 1, 1, ' ', 0)
			_, _ = fmt.Fprintln(w, "WORKFLOW\tSTAGE\tRULES\tACTIONS")
			for _, wf := range cfg.Registry.Workflows {
				for _, stage := range wf.Stages {
					_, _ = fmt.Fprintln(w, wf.Name+"\t"+
						stage.Name+"\t"+
						rules.EvaluateRulesAsText(stage.Rules, rules.GetRuleContext(env))+"\t"+
						strconv.Itoa(len(stage.Actions)))
				}
			}
			_ = w.Flush()
		},
	}

	return cmd
}
