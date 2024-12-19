package cmd

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"sync"

	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cid/pkg/core/rules"
	"github.com/cidverse/cidverseutils/core/clioutputwriter"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/rs/zerolog/log"
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
			format, _ := cmd.Flags().GetString("format")
			workflows, _ := cmd.Flags().GetStringArray("workflow")

			// app context
			cid, err := context.NewAppContext()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to prepare app context")
				os.Exit(1)
			}

			// data
			data := clioutputwriter.TabularData{
				Headers: []string{"WORKFLOW", "STAGE", "RULES", "ACTIONS"},
				Rows:    [][]interface{}{},
			}
			for _, wf := range cid.Config.Registry.Workflows {
				if len(workflows) > 0 && !slices.Contains(workflows, wf.Name) {
					continue
				}

				for _, stage := range wf.Stages {
					data.Rows = append(data.Rows, []interface{}{
						wf.Name,
						stage.Name,
						rules.EvaluateRulesAsText(stage.Rules, rules.GetRuleContext(cid.Env)),
						strconv.Itoa(len(stage.Actions)),
					})
				}
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
	cmd.Flags().StringArrayP("workflow", "w", []string{}, "filter by workflow (default: all workflows)")
	cmd.Flags().StringP("format", "f", string(clioutputwriter.DefaultOutputFormat()), fmt.Sprintf("output format %s", clioutputwriter.SupportedOutputFormats()))

	return cmd
}
