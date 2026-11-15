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
			plan, networkGlobal, err := loadPlan(cid, stateFile, stateWfName, planEnv)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load plan")
				os.Exit(1)
			}

			// egress filter: re-executes this command as a child process under the filter, never returns
			if egressFilterEnabled(cmd) {
				runPlanExecuteWithEgress(cmd, cid.ProjectDir, plan, networkGlobal, stages, steps)
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
	addPlanEgressFlags(cmd)

	return cmd
}

// loadPlan returns the plan from an environment variable (self-contained pipelines), the state file (generated
// pipelines) or generates it from the current project. The workflow's global network allow list is only available
// from the state file.
func loadPlan(cid *context.CIDContext, stateFile string, stateWfName string, planEnv string) (plangenerate.Plan, plangenerate.StepAccessNetwork, error) {
	if planEnv != "" && stateWfName != "" {
		return plangenerate.Plan{}, nil, fmt.Errorf("--plan-env and --state-wf-name are mutually exclusive")
	}
	if stateFile == "" {
		stateFile = filepath.Join(".cid", "state.json")
	}

	if planEnv != "" {
		// plan embedded in the CI configuration (self-contained GitLab pipelines rendered by cid app-server)
		encoded := os.Getenv(planEnv)
		if encoded == "" {
			return plangenerate.Plan{}, nil, fmt.Errorf("plan environment variable %s is empty or not set", planEnv)
		}
		plan, err := plangenerate.DecodePlan(encoded)
		if err != nil {
			return plangenerate.Plan{}, nil, fmt.Errorf("failed to decode plan from environment variable %s: %w", planEnv, err)
		}
		return plan, nil, nil
	}

	if stateWfName != "" {
		state, err := appconfig.ReadWorkflowState(filepath.Join(cid.ProjectDir, stateFile))
		if err != nil {
			return plangenerate.Plan{}, nil, fmt.Errorf("failed to read workflow state file: %w", err)
		}
		// iterate over all workflows to match via workflow key
		for pair := state.Workflows.Newest(); pair != nil; pair = pair.Prev() {
			if pair.Value.WorkflowKey == stateWfName {
				return pair.Value.Plan, pair.Value.NetworkGlobal, nil
			}
		}
		return plangenerate.Plan{}, nil, fmt.Errorf("workflow key %s not found in state file", stateWfName)
	}

	plan, err := plangenerate.GeneratePlan(plangenerate.GeneratePlanRequest{
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
		return plangenerate.Plan{}, nil, fmt.Errorf("failed to generate action plan: %w", err)
	}
	return plan, nil, nil
}
