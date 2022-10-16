package python

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

type BuildActionStruct struct{}

// GetDetails retrieves information about the action
func (action BuildActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "python-build",
		Version:   "0.1.0",
		UsedTools: []string{"pipenv", "pip"},
	}
}

// Check evaluates if the action should be executed or not
func (action BuildActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return true
}

// Execute runs the action
func (action BuildActionStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	if ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemRequirementsTXT {
		command.RunCommand(`pip install -r requirements.txt`, ctx.Env, ctx.ProjectDir)
	} else if ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemPipfile {
		command.RunCommand(`pipenv install`, ctx.Env, ctx.ProjectDir)
	} else if ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemSetupPy {
		command.RunCommand(`pip install `+ctx.ProjectDir, ctx.Env, ctx.ProjectDir)
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}
