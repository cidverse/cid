package python

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

type CheckActionStruct struct{}

// GetDetails retrieves information about the action
func (action CheckActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "sast",
		Name:      "python-lint",
		Version:   "0.1.0",
		UsedTools: []string{"pipenv", "pip"},
	}
}

// Check evaluates if the action should be executed or not
func (action CheckActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectPythonProject(ctx.ProjectDir)
}

// Execute runs the action
func (action CheckActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	return command.RunOptionalCommand(`flake8 .`, ctx.Env, ctx.ProjectDir)
}

func init() {
	api.RegisterBuiltinAction(CheckActionStruct{})
}
