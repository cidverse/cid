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
	return string(config.ActionTypeBuiltinGolang)
}

func (e Executor) Execute(ctx *commonapi.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *config.Action, action *config.WorkflowAction) error {
	// actionType: builtin
	builtinAction := commonapi.BuiltinActions[catalogAction.Name]
	if builtinAction != nil {
		actErr := builtinAction.Execute(ctx, localState)
		if actErr != nil {
			log.Fatal().Err(actErr).Str("action", action.ID).Msg("action execution failed")
		}
	} else {
		log.Error().Str("action", action.ID).Str("executorName", e.GetName()).Str("executorVersion", e.GetVersion()).Str("executorType", e.GetType()).Msg("action is not registered")
	}

	return nil
}
