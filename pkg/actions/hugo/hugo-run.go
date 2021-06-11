package hugo

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

// Action implementation
type RunActionStruct struct{}

// GetDetails returns information about this action
func (action RunActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "run",
		Name:      "hugo-run",
		Version:   "0.1.0",
		UsedTools: []string{"hugo"},
	}
}

// Check if this package can handle the current environment
func (action RunActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectHugoProject(ctx.ProjectDir)
}

// Check if this package can handle the current environment
func (action RunActionStruct) Execute(ctx api.ActionExecutionContext) {
	_ = command.RunOptionalCommand(`hugo server --minify --gc --log --verboseLog --baseUrl "/" --watch --source `+ctx.ProjectDir+``, ctx.Env, ctx.ProjectDir)
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(RunActionStruct{})
}
