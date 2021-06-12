package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

type TestActionStruct struct{}

func (action TestActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:            "test",
		Name:             "golang-test",
		Version:          "0.1.0",
		UsedTools:        []string{"go"},
		ToolDependencies: GetDependencies(ctx.ProjectDir),
	}
}

func (action TestActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectGolangProject(ctx.ProjectDir)
}

func (action TestActionStruct) Execute(ctx api.ActionExecutionContext) {
	command.RunCommand(`go test -cover ./...`, ctx.Env, ctx.ProjectDir)
}

func init() {
	api.RegisterBuiltinAction(TestActionStruct{})
}
