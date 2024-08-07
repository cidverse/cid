package cmd

import (
	"fmt"
	"os"
	"sync"
	"text/tabwriter"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/cidverse/repoanalyzer/analyzer"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func moduleRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "module",
		Aliases: []string{"m"},
		Short:   ``,
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(0)
		},
	}

	cmd.AddCommand(moduleListCmd())

	return cmd
}

func moduleListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "lists all project modules",
		Run: func(cmd *cobra.Command, args []string) {
			zerolog.SetGlobalLevel(zerolog.WarnLevel)

			// find project directory and load config
			projectDir := api.FindProjectDir()

			// analyze
			modules := analyzer.ScanDirectory(projectDir)

			// print list
			w := tabwriter.NewWriter(redact.NewProtectedWriter(nil, os.Stdout, &sync.Mutex{}, nil), 1, 1, 1, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAME\tBUILD-SYSTEM\tBUILD-SYNTAX\tSUBMODULES")
			for _, module := range modules {
				_, _ = fmt.Fprintln(w, module.Name+"\t"+string(module.BuildSystem)+"\t"+string(module.BuildSystemSyntax)+"\t0")
			}
			_ = w.Flush()
		},
	}

	return cmd
}
