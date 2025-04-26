package planexecute

import (
	"fmt"
	"github.com/cidverse/cid/internal/state"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/actionexecutor"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type ExecuteContext struct {
	Cfg           *config.CIDConfig
	Modules       []*analyzerapi.ProjectModule
	Env           map[string]string
	ProjectDir    string
	StagesFilter  []string
	ModulesFilter []string
	StepFilter    []string
}

func RunPlan(plan plangenerate.Plan, planContext ExecuteContext) {
	log.Debug().Str("plan", plan.Name).Strs("stages", plan.Stages).Msg("workflow start")
	start := time.Now()

	if planContext.StepFilter != nil && len(planContext.StepFilter) > 0 {
		// run steps directly, match via id or slug
		for _, step := range plan.Steps {
			if !slices.Contains(planContext.StepFilter, step.ID) && !slices.Contains(planContext.StepFilter, step.Slug) {
				continue
			}

			RunPlanStep(plan, planContext, step)
		}
	} else {
		// run stages
		for _, stageName := range plan.Stages {
			if len(planContext.StagesFilter) == 0 || slices.Contains(planContext.StagesFilter, stageName) {
				RunPlanStage(plan, planContext, stageName)
			} else {
				log.Debug().Str("workflow", plan.Name).Str("stage", stageName).Strs("filter", planContext.StagesFilter).Msg("stage has been skipped")
			}
		}
	}

	log.Info().Str("plan", plan.Name).Str("duration", time.Since(start).String()).Msg("workflow completed")
}

func RunPlanStage(plan plangenerate.Plan, planContext ExecuteContext, stageName string) {
	log.Debug().Str("stage", stageName).Msg("stage start")
	start := time.Now()

	for _, step := range plan.Steps {
		if step.Stage != stageName {
			continue
		}

		RunPlanStep(plan, planContext, step)
	}

	// complete
	log.Info().Str("stage", stageName).Str("duration", time.Since(start).String()).Msg("stage completed")
}

func RunPlanStep(plan plangenerate.Plan, planContext ExecuteContext, step plangenerate.Step) {
	log.Debug().Str("action", step.Name).Msg("action start")
	catalogAction := planContext.Cfg.Registry.FindAction(step.Action)
	if catalogAction == nil {
		log.Fatal().Str("action_id", step.Action).Msg("workflow configuration error, referencing actions that do not exist")
		os.Exit(1)
	}
	actionContext := api.GetActionContext(planContext.Modules, planContext.ProjectDir, planContext.Env, catalogAction.Metadata.Access)
	actionContext.Config = &step.Config

	// set CurrentModule ref for module-scoped actions
	if step.Scope == catalog.ActionScopeModule {
		var moduleRef analyzerapi.ProjectModule
		for _, m := range planContext.Modules {
			if m.ID == step.Module {
				moduleRef = *m
				break
			}
		}

		if moduleRef.ID == "" {
			log.Fatal().Str("module", step.Module).Msg("module not found")
			os.Exit(1)
		}

		actionContext.CurrentModule = &moduleRef
	}

	RunAction(actionContext, catalogAction, step)

	log.Debug().Str("action", step.Name).Msg("action end")
}

func RunAction(actionContext api.ActionExecutionContext, catalogAction *catalog.Action, step plangenerate.Step) {
	start := time.Now()

	currentModule := "root"
	if actionContext.CurrentModule != nil {
		currentModule = actionContext.CurrentModule.Slug
	}
	log.Info().Str("action", step.Name).Str("module", currentModule).Msg("action start")

	// state: retrieve/init
	localState := state.GetStateFromDirectory(actionContext.Paths.Artifact)
	localState.Modules = actionContext.Modules

	// add action to log
	localState.AuditLog = append(localState.AuditLog, state.AuditEvents{
		Timestamp: time.Now(),
		Type:      "action",
		Payload: map[string]string{
			"action": catalogAction.Repository + "/" + catalogAction.Metadata.Name + "@" + catalogAction.Version,
			"uri":    fmt.Sprintf("oci://%s", catalogAction.Container.Image),
		},
	})

	// serialize action config for pass-thru
	actConfig, _ := yaml.Marshal(&actionContext.Config)
	log.Trace().Str("action", step.Name).Str("type", string(catalogAction.Type)).Str("config", string(actConfig)).Msg("action configuration")

	// paths
	_ = os.MkdirAll(actionContext.Paths.Temp, os.ModePerm)
	_ = os.MkdirAll(actionContext.Paths.Artifact, os.ModePerm)

	// execute
	actionExecutor := actionexecutor.FindExecutorByType(string(catalogAction.Type))
	if actionExecutor != nil {
		err := actionExecutor.Execute(&actionContext, &localState, catalogAction, step)
		if err != nil {
			// TODO: handle error
			log.Fatal().Err(err).Str("action", step.Name).Str("duration", time.Since(start).String()).Str("module", currentModule).Msg("action error")
			return
		}
	} else {
		log.Error().Str("action", step.Name).Str("type", string(catalogAction.Type)).Msg("action type is not supported")
	}

	// state: store
	stateFile := filepath.Join(actionContext.Paths.Artifact, step.Slug, "state.json")
	err := state.WriteStateFile(stateFile, localState)
	if err != nil {
		log.Error().Err(err).Str("action", step.Name).Str("duration", time.Since(start).String()).Str("module", currentModule).Msg("failed to write state file")
		os.Exit(1)
	}

	// complete
	log.Info().Str("action", step.Name).Str("duration", time.Since(start).String()).Str("module", currentModule).Msg("action completed")
}
