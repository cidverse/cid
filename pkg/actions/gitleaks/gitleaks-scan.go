package gitleaks

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
)

type ScanStruct struct{}

// GetDetails retrieves information about the action
func (action ScanStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "sast",
		Name:      "gitleaks-scan",
		Version:   "0.1.0",
		UsedTools: []string{"gitleaks"},
	}
}

// Check evaluates if the action should be executed or not
func (action ScanStruct) Check(ctx api.ActionExecutionContext) bool {
	return vcsrepository.GetVCSRepositoryType(ctx.ProjectDir) == "git"
}

// Execute runs the action
func (action ScanStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	_ = command.RunOptionalCommand(`gitleaks --path=. -v --no-git`, ctx.Env, ctx.ProjectDir)
	return nil
}

func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}
