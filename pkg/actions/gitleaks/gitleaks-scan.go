package gitleaks

import (
	"github.com/cidverse/cid/pkg/core/state"
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
)

type ScanStruct struct{}

// GetDetails retrieves information about the action
func (action ScanStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "gitleaks-scan",
		Version:   "0.1.0",
		UsedTools: []string{"gitleaks"},
	}
}

// Check evaluates if the action should be executed or not
func (action ScanStruct) Check(ctx *api.ActionExecutionContext) bool {
	return vcsrepository.GetVCSRepositoryType(ctx.ProjectDir) == "git" && ctx.Env["GITLEAKS_ENABLED"] == "true"
}

// Execute runs the action
func (action ScanStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	var opts []string
	if ctx.Env["CI"] == "true" {
		opts = append(opts, "--redact")
	}

	_ = command.RunOptionalCommand(`gitleaks --path=. -v --no-git `+strings.Join(opts, " "), ctx.Env, ctx.ProjectDir)
	return nil
}

func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}
