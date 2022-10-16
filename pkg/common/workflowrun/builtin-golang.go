package workflowrun

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/rs/zerolog/log"
)

func evaluateActionBuiltinGolang(ctx *api.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *config.Action, action *config.WorkflowAction) bool {
	// actionType: builtin
	builtinAction := api.BuiltinActions[catalogAction.Name]
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

func runActionBuiltinGolang(ctx *api.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *config.Action, action *config.WorkflowAction) error {
	// actionType: builtin
	builtinAction := api.BuiltinActions[catalogAction.Name]
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
