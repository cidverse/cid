package upx

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

// Action implementation
type OptimizeActionStruct struct{}

// GetDetails returns information about this action
func (action OptimizeActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "build",
		Name:      "upx-optimize",
		Version:   "0.1.0",
		UsedTools: []string{"upx"},
	}
}

// Check if this package can handle the current environment
func (action OptimizeActionStruct) Check(ctx api.ActionExecutionContext) bool {
	fullEnv := api.GetFullEnvironment(ctx.ProjectDir)
	return fullEnv["UPX_ENABLED"] == "true"
}

// Check if this package can handle the current environment
func (action OptimizeActionStruct) Execute(ctx api.ActionExecutionContext) {
	command.RunCommand(`upx --lzma `+ctx.ProjectDir+`/`+Config.Paths.Artifact+`/bin/*`, ctx.Env, ctx.ProjectDir)
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(OptimizeActionStruct{})
}
