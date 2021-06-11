package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

type LintActionStruct struct{}

// GetDetails returns information about this action
func (action LintActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "sast",
		Name:      "golang-lint",
		Version:   "0.1.0",
		UsedTools: []string{"golangci-lint"},
	}
}

// Check if this package can handle the current environment
func (action LintActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectGolangProject(ctx.ProjectDir)
}

// Check if this package can handle the current environment
func (action LintActionStruct) Execute(ctx api.ActionExecutionContext) {
	command.RunCommand(`golangci-lint run`, ctx.Env, ctx.ProjectDir)
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(LintActionStruct{})
}
