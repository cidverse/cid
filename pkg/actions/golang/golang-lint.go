package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"path/filepath"
)

type LintActionStruct struct{}

// GetDetails retrieves information about the action
func (action LintActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "golang-lint",
		Version:   "0.1.0",
		UsedTools: []string{"golangci-lint"},
	}
}

// Check evaluates if the action should be executed or not
func (action LintActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return true
}

// Execute runs the action
func (action LintActionStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	// run lint
	if filesystem.FileExists(filepath.Join(ctx.CurrentModule.Directory, ".golangci.yml")) || filesystem.FileExists(filepath.Join(ctx.ProjectDir, ".golangci.yml")) {
		command.RunCommand(`golangci-lint run --sort-results --issues-exit-code 1`, ctx.Env, ctx.CurrentModule.Directory)
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(LintActionStruct{})
}
