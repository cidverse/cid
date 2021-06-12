package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"strings"
)

type RunActionStruct struct{}

func (action RunActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:            "run",
		Name:             "golang-run",
		Version:          "0.1.0",
		UsedTools:        []string{"go"},
		ToolDependencies: GetDependencies(ctx.ProjectDir),
	}
}

func (action RunActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectGolangProject(ctx.ProjectDir)
}

func (action RunActionStruct) Execute(ctx api.ActionExecutionContext) {
	_ = command.RunOptionalCommand(`go run . `+strings.Join(ctx.Args, " "), ctx.Env, ctx.ProjectDir)
}

func init() {
	api.RegisterBuiltinAction(RunActionStruct{})
}
