package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
)

type LintActionStruct struct{}

// GetDetails retrieves information about the action
func (action LintActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "sast",
		Name:      "golang-lint",
		Version:   "0.1.0",
		UsedTools: []string{"golangci-lint"},
	}
}

// Check evaluates if the action should be executed or not
func (action LintActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return ctx.CurrentModule != nil && ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemGoMod
}

// Execute runs the action
func (action LintActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	command.RunCommand(`golangci-lint run`, ctx.Env, ctx.ProjectDir)
	return nil
}

func init() {
	api.RegisterBuiltinAction(LintActionStruct{})
}
