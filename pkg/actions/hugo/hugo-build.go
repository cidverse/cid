package hugo

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

// Action implementation
type BuildActionStruct struct {}

// GetDetails returns information about this action
func (action BuildActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails {
		Stage: "build",
		Name: "hugo-build",
		Version: "0.1.0",
		UsedTools: []string{"hugo"},
	}
}

// Check if this package can handle the current environment
func (action BuildActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectHugoProject(ctx.ProjectDir)
}

// Check if this package can handle the current environment
func (action BuildActionStruct) Execute(ctx api.ActionExecutionContext) {
	command.RunCommand(`hugo --minify --gc --log --verboseLog --source `+ctx.ProjectDir+` --destination `+ ctx.ProjectDir+`/`+ctx.Paths.Artifact, ctx.Env, ctx.ProjectDir)
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}