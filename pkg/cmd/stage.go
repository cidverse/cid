package cmd

import (
	"fmt"
	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/protectoutput"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"text/tabwriter"
)

func init() {
	rootCmd.AddCommand(stageRootCmd)
	stageRootCmd.AddCommand(stageListCmd)
}

var stageRootCmd = &cobra.Command{
	Use:     "stage",
	Aliases: []string{"s"},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

var stageListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "lists all stages",
	Run: func(cmd *cobra.Command, args []string) {
		// find project directory and load config
		projectDir := api.FindProjectDir()
		cfg := app.Load(projectDir)
		env := api.GetCIDEnvironment(projectDir)

		// print list
		w := tabwriter.NewWriter(protectoutput.NewProtectedWriter(nil, os.Stdout), 1, 1, 1, ' ', 0)
		_, _ = fmt.Fprintln(w, "WORKFLOW\tSTAGE\tRULES\tACTIONS")
		for _, wf := range cfg.Workflows {
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
