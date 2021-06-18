package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/config"
	"github.com/cidverse/cid/pkg/repoanalyzer"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/normalizeci/pkg/common"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"time"
)

const DefaultParallelization = 10

// GetExecutionPlan will generate a automatic execution plan based on the project contents
func GetExecutionPlan(projectDir string, workDir string, env map[string]string, stages []config.WorkflowStage) []config.WorkflowStage {
	var executionPlan []config.WorkflowStage

	if stages == nil {
		stages = FindWorkflowStages(projectDir, env)
	}

	// context
	ctx := GetActionContext(projectDir, env, nil)

	// iterate over all stages
	for _, stage := range stages {
		var stageActions []config.WorkflowAction

		// for each module
		for _, module := range ctx.Modules {
			// customize context
			ctx.CurrentModule = module

			// iterate over all actions
			defaultStageActions := FindWorkflowStageActions(stage.Name, ctx)
			for _, action := range defaultStageActions {
				currentAction := action

				// check action activation criteria
				if currentAction.Type == "builtin" {
					if api.BuiltinActions[currentAction.Name].Check(ctx) {
						currentAction.Module = module
						stageActions = append(stageActions, currentAction)
					}
				} else {
					log.Fatal().Str("action", action.Type+"/"+action.Name).Msg("unsupported action type")
				}
			}
		}

		executionPlan = append(executionPlan, config.WorkflowStage{
			Name:    stage.Name,
			Actions: stageActions,
		})
	}

	return executionPlan
}

func GetExecutionPlanStage(projectDir string, workDir string, env map[string]string, stageName string) *config.WorkflowStage {
	executionPlan := GetExecutionPlan(projectDir, workDir, env, []config.WorkflowStage{{Name: stageName}})

	if len(executionPlan) == 1 {
		return &executionPlan[0]
	}
	return nil
}

// RunStageActions runs all actions of a stage
func RunStageActions(stageName string, projectDir string, workDir string, env map[string]string, args []string) {
	start := time.Now()
	log.Info().Str("stage", stageName).Msg("running stage")

	// find stage actions
	stage := GetExecutionPlanStage(projectDir, workDir, env, stageName)

	if len(stage.Actions) == 0 {
		log.Warn().Str("stage", stageName).Msg("no actions available for current stage")
		return
	}
	for _, action := range stage.Actions {
		RunAction(action, projectDir, env, args)
	}

	log.Info().Str("stage", stageName).Str("duration", time.Since(start).String()).Msg("stage completed")
}

// RunAction runs a specific workflow action
func RunAction(action config.WorkflowAction, projectDir string, env map[string]string, args []string) {
	start := time.Now()

	// serialize action config for passthru
	configAsYaml, _ := yaml.Marshal(&action.Config)
	log.Debug().Str("config", string(configAsYaml)).Msg("action specific config")

	// action context
	ctx := GetActionContext(projectDir, env, action.Module)
	ctx.Config = string(configAsYaml)
	log.Info().Str("action", action.Type+"/"+action.Name).Str("module", ctx.CurrentModule.Slug).Msg("running action")

	// ensure that paths exist
	if !filesystem.DirectoryExists(ctx.Paths.Artifact) {
		filesystem.CreateDirectory(ctx.Paths.Artifact)
	}
	if !filesystem.DirectoryExists(ctx.Paths.Temp) {
		filesystem.CreateDirectory(ctx.Paths.Temp)
	}

	// state: retrieve/init
	stateFile := filepath.Join(ctx.Paths.Artifact, "state.json")
	state := api.ActionStateContext{
		Version: 1,
		Modules: ctx.Modules,
	}
	if filesystem.FileExists(stateFile) {
		stateContent, stateContentErr := filesystem.GetFileContent(stateFile)
		if stateContentErr == nil {
			err := json.Unmarshal([]byte(stateContent), &state)
			if err != nil {
				log.Warn().Err(err).Str("file", stateFile).Msg("failed to restore state")
			}
		}
	}

	// handle action execution
	if action.Type == "builtin" {
		// actionType: builtin
		builtinAction := api.BuiltinActions[action.Name]
		if builtinAction != nil {
			actErr := builtinAction.Execute(ctx, &state)
			if actErr != nil {
				log.Fatal().Err(actErr).Str("action", action.Name).Msg("action execution failed")
			}
		} else {
			log.Error().Str("action", action.Name).Msg("skipping action, does not exist")
		}
	} else {
		log.Fatal().Str("action", action.Name).Str("type", action.Type).Msg("type is not supported")
	}

	// state: store
	stateOut, err := json.Marshal(state)
	if err != nil {
		log.Warn().Err(err).Str("file", stateFile).Msg("failed to store state")
	} else {
		_ = filesystem.RemoveFile(stateFile)

		storeErr := filesystem.SaveFileContent(stateFile, string(stateOut))
		if storeErr != nil {
			log.Warn().Err(storeErr).Str("file", stateFile).Msg("failed to store state")
		}
	}

	log.Info().Str("action", action.Type+"/"+action.Name).Str("duration", time.Since(start).String()).Msg("action completed")
}

// GetActionDetails retrieves the details of a WorkflowAction
func GetActionDetails(action config.WorkflowAction, projectDir string, env map[string]string) api.ActionDetails {
	configAsYaml, _ := yaml.Marshal(&action.Config)
	log.Debug().Str("config", string(configAsYaml)).Msg("action specific config")

	if action.Type == "builtin" {
		// actionType: builtin
		builtinAction := api.BuiltinActions[action.Name]
		if builtinAction != nil {
			// context
			ctx := GetActionContext(projectDir, env, action.Module)

			// run action
			return builtinAction.GetDetails(ctx)
		} else {
			log.Error().Str("action", action.Type+"/"+action.Name).Msg("skipping action, does not exist")
		}
	} else {
		log.Fatal().Str("action", action.Type+"/"+action.Name).Msg("type is not supported")
	}

	return api.ActionDetails{}
}

// FindWorkflowStages finds all relevant stages for the current context (branch, tag, ...)
func FindWorkflowStages(projectDir string, env map[string]string) []config.WorkflowStage {
	// cel expr environment
	celConfig, celConfigErr := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("NCI_COMMIT_REF_PATH", decls.String),
			decls.NewVar("NCI_COMMIT_REF_TYPE", decls.String),
			decls.NewVar("NCI_COMMIT_REF_NAME", decls.String),
		),
	)
	if celConfigErr != nil {
		log.Fatal().Err(celConfigErr).Msg("failed to initialize CEL")
	}

	inputData := map[string]interface{}{
		"NCI_COMMIT_REF_PATH": env["NCI_COMMIT_REF_PATH"],
		"NCI_COMMIT_REF_TYPE": env["NCI_COMMIT_REF_TYPE"],
		"NCI_COMMIT_REF_NAME": env["NCI_COMMIT_REF_NAME"],
	}

	var activeStages []config.WorkflowStage
	for _, stage := range config.Config.Stages {
		if len(stage.Rules) > 0 {
			// evaluate rules
			for _, rule := range stage.Rules {
				if len(rule.Expression) > 0 {
					log.Debug().Str("stage", stage.Name).Str("expression", rule.Expression).Msg("checking expression for stage rule")

					// prepare program for evaluation
					ast, issues := celConfig.Compile(rule.Expression)
					if issues != nil && issues.Err() != nil {
						log.Fatal().Err(issues.Err()).Msg("stage rule type error: " + issues.Err().Error())
					}
					prg, prgErr := celConfig.Program(ast)
					if prgErr != nil {
						log.Fatal().Err(prgErr).Msg("program construction error")
					}

					// evaluate
					execOut, _, execErr := prg.Eval(inputData)
					if execErr != nil {
						log.Warn().Err(execErr).Msg("failed to evaluate filter rule")
					}

					// check result
					if execOut.Type() == types.BoolType {
						if fmt.Sprintf("%+v", execOut) == "true" {
							activeStages = append(activeStages, stage)
							break
						}
					} else {
						log.Warn().Str("stage", stage.Name).Str("expression", rule.Expression).Msg("ignoring stage rule expression, does not return a boolean")
					}
				} else {
					log.Warn().Str("stage", stage.Name).Str("rule", fmt.Sprintf("%+v", rule)).Msg("stage rule can't be evaluated")
				}
			}

		} else {
			activeStages = append(activeStages, stage)
		}
	}

	return activeStages
}

func FindWorkflowStageActions(stage string, ctx api.ActionExecutionContext) []config.WorkflowAction {
	var activeActions []config.WorkflowAction

	for _, action := range config.Config.Actions[stage] {
		if action.Type == "builtin" {
			// actionType: builtin
			builtinAction := api.BuiltinActions[action.Name]
			if builtinAction != nil {
				// add
				if builtinAction.Check(ctx) {
					activeActions = append(activeActions, action)
				}
			} else {
				log.Error().Str("action", action.Type+"/"+action.Name).Msg("skipping action, does not exist")
			}
		}
	}

	return activeActions
}

func FindWorkflowAction(search string) (config.WorkflowAction, error) {
	for _, actions := range config.Config.Actions {
		for _, act := range actions {
			// match by type/name or name
			if act.Type+"/"+act.Name == search {
				return act, nil
			} else if act.Name == search {
				return act, nil
			}
		}
	}

	return config.WorkflowAction{}, errors.New("no action found with query " + search)
}

// GetActionContext gets the action context, this operation is expensive and should only be called once per execution
func GetActionContext(projectDir string, env map[string]string, currentModule *analyzerapi.ProjectModule) api.ActionExecutionContext {
	return api.ActionExecutionContext{
		Paths: config.PathConfig{
			Artifact: filepath.Join(projectDir, "dist"),
			Temp:     filepath.Join(projectDir, "tmp"),
			Cache:    "",
		},
		ProjectDir:      projectDir,
		WorkDir:         filesystem.GetWorkingDirectory(),
		Config:          "",
		Args:            nil,
		Env:             env,
		MachineEnv:      common.GetMachineEnvironment(),
		Parallelization: DefaultParallelization,
		Modules:         repoanalyzer.AnalyzeProject(projectDir, filesystem.GetWorkingDirectory()),
		CurrentModule:   currentModule,
	}
}
