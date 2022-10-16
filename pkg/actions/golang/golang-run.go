package golang

import (
	"github.com/cidverse/cid/pkg/core/state"
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

type RunActionStruct struct{}

// GetDetails retrieves information about the action
func (action RunActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:             "golang-run",
		Version:          "0.1.0",
		UsedTools:        []string{"go"},
		ToolDependencies: GetToolDependencies(ctx),
	}
}

// Check evaluates if the action should be executed or not
func (action RunActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return true
}

// Execute runs the action
func (action RunActionStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	_ = command.RunOptionalCommand(`go run . `+strings.Join(ctx.Args, " "), ctx.Env, ctx.ProjectDir)
	return nil
}

func init() {
	api.RegisterBuiltinAction(RunActionStruct{})
}
