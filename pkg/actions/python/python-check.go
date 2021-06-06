package python

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

// Action implementation
type CheckActionStruct struct {}

// GetDetails returns information about this action
func (action CheckActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails {
		Stage: "sast",
		Name: "python-lint",
		Version: "0.1.0",
		UsedTools: []string{"pipenv", "pip"},
	}
}

// Check if this package can handle the current environment
func (action CheckActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectPythonProject(ctx.ProjectDir)
}

// Check if this package can handle the current environment
func (action CheckActionStruct) Execute(ctx api.ActionExecutionContext) {
	command.RunCommand(`flake8 .`, ctx.Env, ctx.ProjectDir)
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(CheckActionStruct{})
}