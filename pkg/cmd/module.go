package cmd

import (
	"fmt"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/protectoutput"
	"github.com/cidverse/cid/pkg/repoanalyzer"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

func init() {
	rootCmd.AddCommand(moduleRootCmd)
	moduleRootCmd.AddCommand(moduleListCmd)
}

var moduleRootCmd = &cobra.Command{
	Use:     "module",
	Aliases: []string{"m"},
	Short:   ``,
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

var moduleListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "lists all project modules",
	Run: func(cmd *cobra.Command, args []string) {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)

		// find project directory and load config
		projectDir := api.FindProjectDir()

		// analyze
		modules := repoanalyzer.AnalyzeProject(projectDir, projectDir)

		// print list
		w := tabwriter.NewWriter(protectoutput.NewProtectedWriter(nil, os.Stdout), 1, 1, 1, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tBUILD-SYSTEM\tBUILD-SYSTEM-SYNTAX\tFILE")
		for _, module := range modules {
			_, _ = fmt.Fprintln(w, module.Name+"\t"+string(module.BuildSystem)+"\t"+string(module.BuildSystemSyntax)+"\t"+module.Discovery)
		}
		_ = w.Flush()
	},
}