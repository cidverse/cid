package app

import (
	"github.com/cidverse/cid/pkg/actions/actgithub"
	"github.com/cidverse/cid/pkg/actions/container"
	"github.com/cidverse/cid/pkg/actions/gitguardian"
	"github.com/cidverse/cid/pkg/actions/gitleaks"
	"github.com/cidverse/cid/pkg/actions/golang"
	"github.com/cidverse/cid/pkg/actions/hugo"
	"github.com/cidverse/cid/pkg/actions/java"
	"github.com/cidverse/cid/pkg/actions/node"
	"github.com/cidverse/cid/pkg/actions/owaspdepcheck"
	"github.com/cidverse/cid/pkg/actions/python"
	"github.com/cidverse/cid/pkg/actions/sonarqube"
	"github.com/cidverse/cid/pkg/actions/upx"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/config"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"time"
)

func Load(projectDirectory string) {
	// load configuration for the current project
	config.LoadConfig(projectDirectory)

	// preload project actions
	GetProjectActions(projectDirectory)

	// dependency detection
	// this will try to discover version constraints from the projects automatically
	dependencyDetectors := [...]map[string]string{
		golang.GetDependencies(projectDirectory),
	}

	for _, dep := range dependencyDetectors {
		for key, version := range dep {
			_, present := config.Config.Dependencies[key]
			if !present {
				config.Config.Dependencies[key] = version
			}
		}
	}
}

var actionCache = make(map[string][]api.ActionStep)

// GetProjectActions returns all supported actions
func GetProjectActions(projectDirectory string) []api.ActionStep {
	_, isPresent := actionCache[projectDirectory]
	if isPresent {
		return actionCache[projectDirectory]
	}


	var actions []api.ActionStep

	actions = append(actions, golang.RunAction())
	actions = append(actions, golang.BuildAction())
	actions = append(actions, golang.TestAction())
	actions = append(actions, golang.LintAction())

	actions = append(actions, java.RunAction())
	actions = append(actions, java.BuildAction())
	actions = append(actions, java.TestAction())
	actions = append(actions, java.PublishAction())

	actions = append(actions, python.BuildAction())
	actions = append(actions, python.RunAction())
	actions = append(actions, python.CheckAction())

	actions = append(actions, node.BuildAction())

	actions = append(actions, hugo.RunAction())
	actions = append(actions, hugo.BuildAction())

	actions = append(actions, upx.OptimizeAction())

	actions = append(actions, container.PackageAction())

	actions = append(actions, gitleaks.ScanAction())
	actions = append(actions, gitguardian.ScanAction())
	actions = append(actions, sonarqube.ScanAction())
	actions = append(actions, owaspdepcheck.ScanAction())

	actionCache[projectDirectory] = actions
	return actions
}

// DiscoverExecutionPlan will generate a automatic execution plan based on the project contents
func DiscoverExecutionPlan(projectDir string, env map[string]string) []config.WorkflowStage {
	var executionPlan []config.WorkflowStage

	// iterate over all stages
	for _, stage := range config.FindWorkflowStages(projectDir, env) {
		var stageActions []config.WorkflowAction

		// iterate over all actions
		for _, action := range GetProjectActions(projectDir) {
			if action.GetDetails(projectDir, env).Stage == stage.Name {
				// add relevant actions to final execution plan
				if action.Check(projectDir, env) {
					stageActions = append(stageActions, config.WorkflowAction{
						Name:   action.GetDetails(projectDir, env).Name,
						Type:   "builtin",
						Config: nil,
					})
				}
			}
		}

		executionPlan = append(executionPlan, config.WorkflowStage{
			Name:   stage.Name,
			Actions: stageActions,
		})
	}

	return executionPlan
}

func FindActionsByStage(stage string, projectDirectory string, env map[string]string) []api.ActionStep {
	var actions []api.ActionStep

	for _, action := range GetProjectActions(projectDirectory) {
		if stage == action.GetDetails(projectDirectory, env).Stage {
			log.Debug().Str("action", action.GetDetails(projectDirectory, env).Name).Msg("checking action conditions")

			if action.Check(projectDirectory, env) {
				actions = append(actions, action)
			} else {
				log.Debug().Str("action", action.GetDetails(projectDirectory, env).Name).Msg("check failed")
			}
		}
	}

	return actions
}

func FindActionByName(name string, projectDirectory string, env map[string]string) api.ActionStep {
	for _, action := range GetProjectActions(projectDirectory) {
		if name == action.GetDetails(projectDirectory, env).Name {
			return action
		}
	}

	return nil
}

func RunStageActions(stage string, projectDirectory string, env map[string]string, args []string) {
	start := time.Now()

	if config.Config.Actions != nil && len(config.Config.Actions[stage]) > 0 {
		// custom actions
		for _, currentAction := range config.Config.Actions[stage] {
			RunAction(currentAction, projectDirectory, env, args)
		}
	} else {
		// auto-detected actions
		actions := FindActionsByStage(stage, projectDirectory, env)
		if len(actions) == 0 {
			log.Fatal().Str("projectDirectory", projectDirectory).Msg("can't detect project type")
		}
		for _, action := range actions {
			RunAction(config.WorkflowAction{Name: action.GetDetails(projectDirectory, env).Name}, projectDirectory, env, args)
		}
	}

	log.Info().Str("stage", stage).Str("duration", time.Now().Sub(start).String()).Msg("stage complete")
}

func RunAction(action config.WorkflowAction, projectDirectory string, env map[string]string, args []string) {
	if len(action.Type) == 0 {
		action.Type = "builtin"
	}
	log.Info().Str("action", action.Name).Str("actionType", action.Type).Msg("running action")

	configAsYaml, _ := yaml.Marshal(&action.Config)
	log.Debug().Str("config", string(configAsYaml)).Msg("action specific config")

	if action.Type == "builtin" {
		// actionType: builtin
		builtinAction := FindActionByName(action.Name, projectDirectory, env)
		if builtinAction != nil {
			// pass config
			builtinAction.SetConfig(string(configAsYaml))

			// run action
			builtinAction.Execute(projectDirectory, env, args)
		} else {
			log.Error().Str("action", action.Name).Msg("skipping action, does not exist")
		}
	} else {
		log.Error().Str("action", action.Name).Str("type", action.Type).Msg("type is not supported")
	}
}
