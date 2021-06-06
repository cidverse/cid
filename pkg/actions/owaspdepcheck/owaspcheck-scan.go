package owaspdepcheck

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

// Action implementation
type ScanStruct struct {}

// GetDetails returns information about this action
func (action ScanStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails {
		Stage: "sast",
		Name: "dependencycheck-scan",
		Version: "0.0.1",
		UsedTools: []string{"dependency-check"},
	}
}

// Check if this package can handle the current environment
func (action ScanStruct) Check(ctx api.ActionExecutionContext) bool {
	return true
}

// Check if this package can handle the current environment
func (action ScanStruct) Execute(ctx api.ActionExecutionContext) {
	_ = command.RunOptionalCommand(`dependency-check --noupdate --scan . --enableExperimental --out dist --exclude .git/** --exclude dist/**`, ctx.Env, ctx.ProjectDir)
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}