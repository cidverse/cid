package api

import (
	commonapi "github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/state"
)

// ActionExecutor is the interface that needs to be implemented by all action executors
type ActionExecutor interface {
	// GetName returns the name of the executor
	GetName() string

	// GetVersion returns the version of the executor
	GetVersion() string

	// GetType returns the action type which needs to match the config action type to activate this implementation
	GetType() string

	// Check will evaluate if the action should be executed
	Check(ctx *commonapi.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *config.Action, action *config.WorkflowAction) bool

	// Execute will run the action
	Execute(ctx *commonapi.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *config.Action, action *config.WorkflowAction) error
}
