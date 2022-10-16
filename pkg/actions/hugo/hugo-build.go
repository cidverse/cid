package hugo

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/state"
)

type BuildActionStruct struct{}

// GetDetails retrieves information about the action
func (action BuildActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "hugo-build",
		Version:   "0.1.0",
		UsedTools: []string{"hugo"},
	}
}

// Check evaluates if the action should be executed or not
func (action BuildActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return true
}

// Execute runs the action
func (action BuildActionStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	command.RunCommand(`hugo --minify --gc --log --verboseLog --source `+ctx.ProjectDir+` --destination `+ctx.ProjectDir+`/`+ctx.Paths.Artifact, ctx.Env, ctx.ProjectDir)

	return nil
}

func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}
