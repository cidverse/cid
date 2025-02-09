package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cidverseutils/core/clioutputwriter"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func executeRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "execute",
		Aliases: []string{},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(0)
		},
	}

	cmd.AddCommand(executableCandidateListCmd())

	return cmd
}

func executableCandidateListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-candidates",
		Aliases: []string{},
		Short:   "list all executable candidates",
		Run: func(cmd *cobra.Command, args []string) {
			format, _ := cmd.Flags().GetString("format")

			// app context
			cid, err := context.NewAppContext()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to prepare app context")
				os.Exit(1)
			}

			// get candidates
			candidates, err := command.CandidatesFromConfig(*cid.Config)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to discover candidates")
				os.Exit(1)
			}

			// data
			data := clioutputwriter.TabularData{
				Headers: []string{"NAME", "TYPE", "VERSION", "URI"},
				Rows:    [][]interface{}{},
			}
			for _, candidate := range candidates {
				data.Rows = append(data.Rows, []interface{}{
					candidate.GetName(),
					string(candidate.GetType()),
					candidate.GetVersion(),
					candidate.GetUri(),
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
	cmd.Flags().StringP("format", "f", string(clioutputwriter.DefaultOutputFormat()), fmt.Sprintf("output format %s", clioutputwriter.SupportedOutputFormats()))
	return cmd
}
