package cmd

import (
	"os"
	"strconv"
	"sync"

	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cid/pkg/core/cmdoutput"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/cidverse/repoanalyzer/analyzer"
	"github.com/rs/zerolog/log"
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
			format, _ := cmd.Flags().GetString("format")

			// app context
			cid, err := context.NewAppContext()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to prepare app context")
				os.Exit(1)
			}

			// analyze
			modules := analyzer.ScanDirectory(cid.ProjectDir)

			// data
			data := cmdoutput.TabularData{
				Headers: []string{"NAME", "TYPE", "BUILD-SYSTEM", "BUILD-SYNTAX", "SPEC-TYPE", "SUBMODULES"},
				Rows:    [][]string{},
			}
			for _, module := range modules {
				data.Rows = append(data.Rows, []string{
					module.Name,
					string(module.Type),
					string(module.BuildSystem),
					string(module.BuildSystemSyntax),
					string(module.SpecificationType),
					strconv.Itoa(len(module.Submodules)),
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
