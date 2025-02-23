package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/repoanalyzer/analyzer"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func planRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plan",
		Aliases: []string{},
		Short:   ``,
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(0)
		},
	}

	cmd.AddCommand(planGenerateCmd())

	return cmd
}

func planGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate",
		Aliases: []string{"gen"},
		Short:   "",
		Run: func(cmd *cobra.Command, args []string) {
			// app context
			cid, err := context.NewAppContext()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to prepare app context")
				os.Exit(1)
			}

			// analyze
			modules := analyzer.ScanDirectory(cid.ProjectDir)

			// data
			plan, err := plangenerate.GeneratePlan(modules, cid.Config.Registry, cid.ProjectDir, cid.Env)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to generate action plan")
				os.Exit(1)
			}

			// output
			out, _ := json.MarshalIndent(plan, "", "  ")
			fmt.Println(string(out))
		},
	}

	return cmd
}
