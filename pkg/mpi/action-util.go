package mpi

import (
	"github.com/PhilippHeuer/cid/pkg/actions/golang"
	"github.com/PhilippHeuer/cid/pkg/actions/hugo"
	"github.com/PhilippHeuer/cid/pkg/actions/java"
	"github.com/PhilippHeuer/cid/pkg/actions/upx"
	"github.com/PhilippHeuer/cid/pkg/common/api"
	"github.com/rs/zerolog/log"
)

// GetName returns the name
func GetAllActions() []api.ActionStep {
	var actions []api.ActionStep
	actions = append(actions, golang.RunAction())
	actions = append(actions, golang.BuildAction())
	actions = append(actions, golang.TestAction())
	actions = append(actions, java.BuildAction())
	actions = append(actions, hugo.RunAction())
	actions = append(actions, hugo.BuildAction())
	actions = append(actions, upx.OptimizeAction())

	return actions
}

func FindActionByStage(stage string, projectDir string) []api.ActionStep {
	var actions []api.ActionStep

	for _, action := range GetAllActions() {
		if stage == action.GetStage() {
			log.Debug().Str("action", action.GetName()).Msg("checking action conditions")

			if action.Check(projectDir) {
				actions = append(actions, action)
			} else {
				log.Debug().Str("action", action.GetName()).Msg("check failed")
			}
		}
	}

	return actions
}

func FindActionByName(name string) api.ActionStep {
	for _, action := range GetAllActions() {
		if name == action.GetName() {
			return action
		}
	}

	return nil
}

func RunStageActions(stage string, projectDirectory string, ciEnv []string, args []string) {
	// load workflow config
	loadConfig(projectDirectory)
	finalEnv := api.GetEffectiveEnv(ciEnv)

	if Config.Workflow != nil && len(Config.Workflow) > 0 {
		for _, currentStage := range Config.Workflow {
			if currentStage.Stage == stage {
				if len(currentStage.Actions) > 0 {
					for _, currentAction := range currentStage.Actions {
						action := FindActionByName(currentAction.Name)
						if action != nil {
							action.Execute(projectDirectory, finalEnv, args)
						} else {
							log.Error().Str("action", currentAction.Name).Msg("skipping action, does not exist")
						}
					}

					return
				} else {
					// stage configuration present but no actions configured
				}
			} else {
				// not custom actions configured for this stage
			}
		}
	}

	// actions
	actions := FindActionByStage(stage, projectDirectory)
	if len(actions) == 0 {
		log.Fatal().Str("projectDirectory", projectDirectory).Msg("can't detect project type")
	}
	for _, action := range actions {
		action.Execute(projectDirectory, finalEnv, args)
	}
}
