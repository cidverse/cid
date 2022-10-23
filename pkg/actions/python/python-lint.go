package python

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/state"
)

type LintActionStruct struct{}

// GetDetails retrieves information about the action
func (action LintActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "python-lint",
		Version:   "0.1.0",
		UsedTools: []string{"pipenv", "pip"},
	}
}

// Execute runs the action
func (action LintActionStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	return command.RunOptionalCommand(`flake8 .`, ctx.Env, ctx.ProjectDir)
}

func init() {
	api.RegisterBuiltinAction(LintActionStruct{})
}
