package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
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
	return ctx.CurrentModule != nil && ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemGoMod
}

// Execute runs the action
func (action LintActionStruct) Execute(ctx *api.ActionExecutionContext, state *api.ActionStateContext) error {
	// golangci lint preset
	configFile := filepath.Join(ctx.CurrentModule.Directory, ".golangci.yml")
	if !filesystem.FileExists(configFile) {
		content, _ := embeddedConfigFS.ReadFile("files/golangci-default.yml")
		_ = filesystem.CreateFileWithContent(configFile, string(content))
	}

	// run lint
	command.RunCommand(`golangci-lint run --sort-results --issues-exit-code 1`, ctx.Env, ctx.CurrentModule.Directory)

	return nil
}

func init() {
	api.RegisterBuiltinAction(LintActionStruct{})
}
