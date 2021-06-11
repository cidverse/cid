package python

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

// Action implementation
type BuildActionStruct struct{}

// GetDetails returns information about this action
func (action BuildActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "build",
		Name:      "python-build",
		Version:   "0.1.0",
		UsedTools: []string{"pipenv", "pip"},
	}
}

// Check will evaluate if this action can be executed for the specified project
func (action BuildActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectPythonProject(ctx.ProjectDir)
}

// Execute will run the action
func (action BuildActionStruct) Execute(ctx api.ActionExecutionContext) {
	buildSystem := DetectPythonBuildSystem(ctx.ProjectDir)
	if buildSystem == "requirements.txt" {
		command.RunCommand(`pip install -r requirements.txt`, ctx.Env, ctx.ProjectDir)
	} else if buildSystem == "pipenv" {
		command.RunCommand(`pipenv install`, ctx.Env, ctx.ProjectDir)
	} else if buildSystem == "setup.py" {
		command.RunCommand(`pip install `+ctx.ProjectDir, ctx.Env, ctx.ProjectDir)
	}
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}
