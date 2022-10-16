package hugo

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/state"
)

type RunActionStruct struct{}

// GetDetails retrieves information about the action
func (action RunActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "hugo-run",
		Version:   "0.1.0",
		UsedTools: []string{"hugo"},
	}
}

// Check evaluates if the action should be executed or not
func (action RunActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return true
}

// Execute runs the action
func (action RunActionStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	_ = command.RunOptionalCommand(`hugo server --minify --gc --log --verboseLog --baseUrl "/" --watch --source `+ctx.ProjectDir+``, ctx.Env, ctx.ProjectDir)

	return nil
}

func init() {
	api.RegisterBuiltinAction(RunActionStruct{})
}
