package gitleaks

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
)

// Action implementation
type ScanStruct struct {}

// GetDetails returns information about this action
func (action ScanStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails {
		Stage: "sast",
		Name: "gitleaks-scan",
		Version: "0.1.0",
		UsedTools: []string{"gitleaks"},
	}
}

// Check if this package can handle the current environment
func (action ScanStruct) Check(ctx api.ActionExecutionContext) bool {
	return vcsrepository.GetVCSRepositoryType(ctx.ProjectDir) == "git"
}

// Check if this package can handle the current environment
func (action ScanStruct) Execute(ctx api.ActionExecutionContext) {
	_ = command.RunOptionalCommand(`gitleaks --path=. -v --no-git`, ctx.Env, ctx.ProjectDir)
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}