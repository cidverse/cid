package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cidverse/cid/pkg/app/appconfig"
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
				Modules:      cid.Modules,
				Registry:     cid.Config.Registry,
				ProjectDir:   cid.ProjectDir,
				Env:          cid.Env,
				Executables:  cid.Executables,
				PinVersions:  pin,
				WorkflowType: "",
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
			stateFile, _ := cmd.Flags().GetString("state-file")
			stateWfName, _ := cmd.Flags().GetString("state-wf-name")
			planEnv, _ := cmd.Flags().GetString("plan-env")
			if planEnv != "" && stateWfName != "" {
				slog.Error("--plan-env and --state-wf-name are mutually exclusive")
				os.Exit(1)
			}
			if stateFile == "" {
				stateFile = filepath.Join(".cid", "state.json")
			}

			// app context
			cid, err := context.NewAppContext()
			if err != nil {
				slog.With("err", err).Error("failed to prepare app context")
				os.Exit(1)
			}

			// read plan file
			var plan plangenerate.Plan
			if planEnv != "" {
				// plan embedded in the CI configuration (self-contained GitLab pipelines rendered by cid app-server)
				encoded := os.Getenv(planEnv)
				if encoded == "" {
					slog.With("env", planEnv).Error("plan environment variable is empty or not set")
					os.Exit(1)
				}
				plan, err = plangenerate.DecodePlan(encoded)
				if err != nil {
					slog.With("env", planEnv).With("err", err).Error("failed to decode plan from environment variable")
					os.Exit(1)
				}
			} else if stateWfName != "" {
				state, stateReadErr := appconfig.ReadWorkflowState(filepath.Join(cid.ProjectDir, stateFile))
				if stateReadErr != nil {
					log.Fatal().Err(stateReadErr).Msg("failed to read workflow state file")
					os.Exit(1)
				}

				// iterate over all workflows to match via workflow key
				for pair := state.Workflows.Newest(); pair != nil; pair = pair.Prev() {
					if pair.Value.WorkflowKey == stateWfName {
						plan = pair.Value.Plan
						break
					}
				}
				if plan.Name == "" {
					slog.With("workflow_key", stateWfName).Error("workflow key not found in state file")
					os.Exit(1)
				}
			} else {
				// generate
				plan, err = plangenerate.GeneratePlan(plangenerate.GeneratePlanRequest{
					Modules:      cid.Modules,
					Registry:     cid.Config.Registry,
					ProjectDir:   cid.ProjectDir,
					Env:          cid.Env,
					Executables:  cid.Executables,
					PinVersions:  false,
					Environments: nil,
					WorkflowType: "",
				})
				if err != nil {
					log.Fatal().Err(err).Msg("failed to generate action plan")
					os.Exit(1)
				}
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
	cmd.Flags().String("state-file", "", "path to the state file, defaults to .cid/state.json")
	cmd.Flags().String("state-wf-name", "", "workflow name, MUST BE present in .cid/state.json")
	cmd.Flags().String("plan-env", "", "name of an environment variable holding the plan as base64(gzip(json)), used by self-contained pipelines instead of --state-file / --state-wf-name")

	return cmd
}
