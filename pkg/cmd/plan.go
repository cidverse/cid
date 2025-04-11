package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cid/pkg/core/planexecute"
	"github.com/cidverse/cid/pkg/core/plangenerate"
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
	cmd.AddCommand(planExecuteCmd())

	return cmd
}

func planGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate",
		Aliases: []string{"g"},
		Short:   "",
		Run: func(cmd *cobra.Command, args []string) {
			pin, _ := cmd.Flags().GetBool("pin")

			// app context
			cid, err := context.NewAppContext()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to prepare app context")
				os.Exit(1)
			}

			// data
			plan, err := plangenerate.GeneratePlan(plangenerate.GeneratePlanRequest{
				Modules:         cid.Modules,
				Registry:        cid.Config.Registry,
				ProjectDir:      cid.ProjectDir,
				Env:             cid.Env,
				Executables:     cid.Executables,
				PinVersions:     pin,
				WorkflowVariant: "",
			})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to generate action plan")
				os.Exit(1)
			}

			// output
			buffer := &bytes.Buffer{}
			encoder := json.NewEncoder(buffer)
			encoder.SetIndent("", "  ")
			encoder.SetEscapeHTML(false)
			_ = encoder.Encode(plan)
			fmt.Println(buffer.String())
		},
	}

	cmd.Flags().Bool("pin", false, "pin all versions when generating the plan")

	return cmd
}

func planExecuteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "execute",
		Aliases: []string{},
		Short:   "",
		Run: func(cmd *cobra.Command, args []string) {
			stages, _ := cmd.Flags().GetStringArray("stage")
			steps, _ := cmd.Flags().GetStringArray("step")

			// app context
			cid, err := context.NewAppContext()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to prepare app context")
				os.Exit(1)
			}

			// TODO: generate plan or take from input file

			// plan
			plan, err := plangenerate.GeneratePlan(plangenerate.GeneratePlanRequest{
				Modules:         cid.Modules,
				Registry:        cid.Config.Registry,
				ProjectDir:      cid.ProjectDir,
				Env:             cid.Env,
				Executables:     cid.Executables,
				PinVersions:     false,
				WorkflowVariant: "",
			})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to generate action plan")
				os.Exit(1)
			}

			// run plan
			planexecute.RunPlan(plan, planexecute.ExecuteContext{
				Cfg:           cid.Config,
				Modules:       cid.Modules,
				Env:           cid.Env,
				ProjectDir:    cid.ProjectDir,
				StagesFilter:  stages,
				ModulesFilter: []string{},
				StepFilter:    steps,
			})
		},
	}

	cmd.Flags().StringArrayP("stage", "s", []string{}, "limit execution to the specified stage(s)")
	cmd.Flags().StringArray("step", []string{}, "limit execution to the specified step(s)")

	return cmd
}
