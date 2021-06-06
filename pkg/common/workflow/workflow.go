package workflow

import (
	"errors"
	"fmt"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/config"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/normalizeci/pkg/common"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"time"
)

// DiscoverExecutionPlan will generate a automatic execution plan based on the project contents
func DiscoverExecutionPlan(projectDir string, env map[string]string) []config.WorkflowStage {
	var executionPlan []config.WorkflowStage

	// context
	ctx := api.ActionExecutionContext{
		Paths: config.Config.Paths,
		ProjectDir: projectDir,
		WorkDir: filesystem.GetWorkingDirectory(),
		Config: "",
		Args: nil,
		Env: env,
		MachineEnv: common.GetMachineEnvironment(),
	}

	// iterate over all stages
	for _, stage := range FindWorkflowStages(projectDir, env) {
		var stageActions []config.WorkflowAction

		// iterate over all actions
		for _, action := range FindWorkflowStageActions(projectDir, env, stage.Name) {
			// check action activation criteria
			if action.Type == "builtin" {
				if api.BuiltinActions[action.Name].Check(ctx) {
					stageActions = append(stageActions, action)
				}
			} else {
				log.Fatal().Str("action", action.Type+"/"+action.Name).Msg("unsupported action type")
			}
		}

		executionPlan = append(executionPlan, config.WorkflowStage{
			Name:   stage.Name,
			Actions: stageActions,
		})
	}

	return executionPlan
}

// RunStageActions runs all actions of a stage
func RunStageActions(stage string, projectDirectory string, env map[string]string, args []string) {
	start := time.Now()
	log.Info().Str("stage", stage).Msg("running stage")

	// find stage actions
	actions := FindWorkflowStageActions(projectDirectory, env, stage)
	if len(actions) == 0 {
		log.Warn().Str("stage", stage).Msg("no actions available for current stage")
		return
	}
	for _, action := range actions {
		RunAction(action, projectDirectory, env, args)
	}

	log.Info().Str("stage", stage).Str("duration", time.Now().Sub(start).String()).Msg("stage completed")
}

// RunAction runs a specific stage
func RunAction(action config.WorkflowAction, projectDirectory string, env map[string]string, args []string) {
	start := time.Now()
	log.Info().Str("action", action.Type+"/"+action.Name).Msg("running action")

	configAsYaml, _ := yaml.Marshal(&action.Config)
	log.Debug().Str("config", string(configAsYaml)).Msg("action specific config")

	if action.Type == "builtin" {
		// context
		ctx := api.ActionExecutionContext{
			Paths: config.Config.Paths,
			ProjectDir: projectDirectory,
			WorkDir: filesystem.GetWorkingDirectory(),
			Config: string(configAsYaml),
			Args: args,
			Env: env,
			MachineEnv: common.GetMachineEnvironment(),
		}

		// actionType: builtin
		builtinAction := api.BuiltinActions[action.Name]
		if builtinAction != nil {
			builtinAction.Execute(ctx)
		} else {
			log.Error().Str("action", action.Name).Msg("skipping action, does not exist")
		}
	} else {
		log.Fatal().Str("action", action.Name).Str("type", action.Type).Msg("type is not supported")
	}

	log.Info().Str("action", action.Type+"/"+action.Name).Str("duration", time.Now().Sub(start).String()).Msg("action completed")
}

// GetActionDetails retrieves the details of a WorkflowAction
func GetActionDetails(action config.WorkflowAction, projectDirectory string, env map[string]string) api.ActionDetails {
	configAsYaml, _ := yaml.Marshal(&action.Config)
	log.Debug().Str("config", string(configAsYaml)).Msg("action specific config")

	if action.Type == "builtin" {
		// actionType: builtin
		builtinAction := api.BuiltinActions[action.Name]
		if builtinAction != nil {
			// context
			ctx := api.ActionExecutionContext{
				Paths: config.Config.Paths,
				ProjectDir: projectDirectory,
				WorkDir: filesystem.GetWorkingDirectory(),
				Config: string(configAsYaml),
				Args: nil,
				Env: env,
				MachineEnv: common.GetMachineEnvironment(),
			}

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
		),
	)
	if celConfigErr != nil {
		log.Fatal().Err(celConfigErr).Msg("failed to initialize CEL")
	}

	inputData := map[string]interface{}{
		"NCI_COMMIT_REF_PATH": env["NCI_COMMIT_REF_PATH"],
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

func FindWorkflowStageActions(projectDir string, env map[string]string, stage string) []config.WorkflowAction {
	var activeActions []config.WorkflowAction

	for _, action := range config.Config.Actions[stage] {

		if action.Type == "builtin" {
			// actionType: builtin
			builtinAction := api.BuiltinActions[action.Name]
			if builtinAction != nil {
				// context
				ctx := api.ActionExecutionContext{
					Paths:      config.Config.Paths,
					ProjectDir: projectDir,
					WorkDir:    filesystem.GetWorkingDirectory(),
					Config:     "",
					Args:       nil,
					Env:        env,
					MachineEnv: common.GetMachineEnvironment(),
				}

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