package python

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

type BuildActionStruct struct{}

// GetDetails retrieves information about the action
func (action BuildActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "python-build",
		Version:   "0.1.0",
		UsedTools: []string{"pipenv", "pip"},
	}
}

// Check evaluates if the action should be executed or not
func (action BuildActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectPythonProject(ctx.ProjectDir)
}

// Execute runs the action
func (action BuildActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	buildSystem := DetectPythonBuildSystem(ctx.ProjectDir)
	if buildSystem == "requirements.txt" {
		command.RunCommand(`pip install -r requirements.txt`, ctx.Env, ctx.ProjectDir)
	} else if buildSystem == "pipenv" {
		command.RunCommand(`pipenv install`, ctx.Env, ctx.ProjectDir)
	} else if buildSystem == "setup.py" {
		command.RunCommand(`pip install `+ctx.ProjectDir, ctx.Env, ctx.ProjectDir)
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}
