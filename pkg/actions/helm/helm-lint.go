package helm

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

type LintActionStruct struct{}

// GetDetails retrieves information about the action
func (action LintActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "helm-lint",
		Version:   "0.1.0",
		UsedTools: []string{"helm"},
	}
}

// Check evaluates if the action should be executed or not
func (action LintActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return true
}

// Execute runs the action
func (action LintActionStruct) Execute(ctx *api.ActionExecutionContext, state *api.ActionStateContext) error {
	command.RunCommand("helm lint "+ctx.CurrentModule.Directory+" --strict", ctx.Env, ctx.ProjectDir)

	return nil
}

func init() {
	api.RegisterBuiltinAction(LintActionStruct{})
}
