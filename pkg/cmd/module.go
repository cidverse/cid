package cmd

import (
	"os"
	"strconv"
	"sync"

	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cidverseutils/core/clioutputwriter"
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
			data := clioutputwriter.TabularData{
				Headers: []string{"NAME", "TYPE", "BUILD-SYSTEM", "BUILD-SYNTAX", "SPEC-TYPE", "SUBMODULES"},
				Rows:    [][]interface{}{},
			}
			for _, module := range modules {
				data.Rows = append(data.Rows, []interface{}{
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
