package builtin

import (
	commonapi "github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/rs/zerolog/log"
)

type Executor struct{}

func (e Executor) GetName() string {
	return "builtin"
}

func (e Executor) GetVersion() string {
	return "0.1.0"
}

func (e Executor) GetType() string {
	return "builtin-golang"
}

func (e Executor) Check(ctx *commonapi.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *config.Action, action *config.WorkflowAction) bool {
	// actionType: builtin
	builtinAction := commonapi.BuiltinActions[catalogAction.Name]
	if builtinAction != nil {
		// runtime check
		if builtinAction.Check(ctx) {
			return true
		}
	} else {
		log.Error().Str("action", action.ID).Msg("skipping action, does not exist")
	}

	return false
}

func (e Executor) Execute(ctx *commonapi.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *config.Action, action *config.WorkflowAction) error {
	// actionType: builtin
	builtinAction := commonapi.BuiltinActions[catalogAction.Name]
	if builtinAction != nil {
		// runtime check
		if builtinAction.Check(ctx) {
			actErr := builtinAction.Execute(ctx, localState)
			if actErr != nil {
				log.Fatal().Err(actErr).Str("action", action.ID).Msg("action execution failed")
			}
		} else {
			log.Warn().Msg("action requirements not fulfilled!")
		}
	} else {
		log.Error().Str("action", action.ID).Msg("skipping action, does not exist")
	}

	return nil
}
