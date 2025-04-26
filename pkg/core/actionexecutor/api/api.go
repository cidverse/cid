package api

import (
	"github.com/cidverse/cid/internal/state"
	commonapi "github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/plangenerate"
)

// ActionExecutor is the interface that needs to be implemented by all action executors
type ActionExecutor interface {
	// GetName returns the name of the executor
	GetName() string

	// GetVersion returns the version of the executor
	GetVersion() string

	// GetType returns the action type which needs to match the config action type to activate this implementation
	GetType() string

	// Execute will run the action
	Execute(ctx *commonapi.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *catalog.Action, step plangenerate.Step) error
}
