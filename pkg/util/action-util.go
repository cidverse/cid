package util

import (
	"github.com/PhilippHeuer/cid/pkg/actions/golang"
	"github.com/PhilippHeuer/cid/pkg/actions/hugo"
	"github.com/PhilippHeuer/cid/pkg/actions/upx"
	"github.com/PhilippHeuer/cid/pkg/common/api"
	"github.com/rs/zerolog/log"
)

// GetName returns the name
func GetAllActions() []api.ActionStep {
	var actions []api.ActionStep
	actions = append(actions, golang.BuildAction())
	actions = append(actions, golang.TestAction())
	actions = append(actions, hugo.BuildAction())
	actions = append(actions, upx.OptimizeAction())

	return actions
}

func FindActionByStage(stage string, projectDir string) api.ActionStep {
	for _, action := range GetAllActions() {
		if stage == action.GetStage() {
			log.Debug().Str("action", action.GetName()).Msg("checking action conditions")

			if action.Check(projectDir) {
				return action
			} else {
				log.Debug().Str("action", action.GetName()).Msg("check failed")
			}
		}
	}

	return nil
}

func FindActionByName(name string) api.ActionStep {
	for _, action := range GetAllActions() {
		if name == action.GetName() {
			return action
		}
	}

	return nil
}
