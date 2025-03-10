package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cidverseutils/core/clioutputwriter"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func executablesRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "executables",
		Aliases: []string{},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(0)
		},
	}

	cmd.AddCommand(executablesUpdateCmd())
	cmd.AddCommand(executablesClearCmd())
	cmd.AddCommand(executablesListCmd())

	return cmd
}

func executablesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "update the candidate-lock.json file",
		Aliases: []string{},
		Run: func(cmd *cobra.Command, args []string) {
			// discover executables
			executables, err := executable.DiscoverExecutables()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to update executables")
				os.Exit(1)
			}

			// update cache
			err = executable.UpdateExecutableCache(executables)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to generate candidate-lock.json")
				os.Exit(1)
			}
		},
	}

	return cmd
}

func executablesClearCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clear",
		Short:   "clears the executable candidate cache",
		Aliases: []string{},
		Run: func(cmd *cobra.Command, args []string) {
			executable.ResetExecutableCache()
		},
	}

	return cmd
}

func executablesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{},
		Short:   "list all executable candidates",
		Run: func(cmd *cobra.Command, args []string) {
			format, _ := cmd.Flags().GetString("format")
			columns, _ := cmd.Flags().GetStringSlice("columns")

			// app context
			cid, err := context.NewAppContext()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to prepare app context")
				os.Exit(1)
			}

			// data
			data := clioutputwriter.TabularData{
				Headers: []string{"NAME", "TYPE", "VERSION", "URI"},
				Rows:    [][]interface{}{},
			}
			for _, c := range cid.Executables {
				data.Rows = append(data.Rows, []interface{}{
					c.GetName(),
					string(c.GetType()),
					c.GetVersion(),
					c.GetUri(),
				})
			}

			// filter columns
			if len(columns) > 0 {
				data = clioutputwriter.FilterColumns(data, columns)
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

	cmd.Flags().StringP("format", "f", string(clioutputwriter.DefaultOutputFormat()), fmt.Sprintf("output format %s", clioutputwriter.SupportedOutputFormats()))
	cmd.Flags().StringSliceP("columns", "c", []string{}, "columns to display")

	return cmd
}
