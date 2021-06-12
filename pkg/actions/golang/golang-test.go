package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

type TestActionStruct struct{}

// GetDetails retrieves information about the action
func (action TestActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:            "test",
		Name:             "golang-test",
		Version:          "0.1.0",
		UsedTools:        []string{"go"},
		ToolDependencies: GetDependencies(ctx.ProjectDir),
	}
}

// Check evaluates if the action should be executed or not
func (action TestActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectGolangProject(ctx.ProjectDir)
}

// Execute runs the action
func (action TestActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	command.RunCommand(`go test -cover ./...`, ctx.Env, ctx.ProjectDir)

	return nil
}

func init() {
	api.RegisterBuiltinAction(TestActionStruct{})
}
