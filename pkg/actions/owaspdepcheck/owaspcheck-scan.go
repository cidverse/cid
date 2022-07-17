package owaspdepcheck

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

type ScanStruct struct{}

// GetDetails retrieves information about the action
func (action ScanStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "dependencycheck-scan",
		Version:   "0.0.1",
		UsedTools: []string{"dependency-check"},
	}
}

// Check evaluates if the action should be executed or not
func (action ScanStruct) Check(ctx *api.ActionExecutionContext) bool {
	return ctx.Env["OWASP_DEPENDENCYCHECK_ENABLED"] == "true"
}

// Execute runs the action
func (action ScanStruct) Execute(ctx *api.ActionExecutionContext, state *api.ActionStateContext) error {
	_ = command.RunOptionalCommand(`dependency-check --noupdate --scan . --enableExperimental --out dist --exclude .git/** --exclude `+ctx.Paths.Artifact+`/**`, ctx.Env, ctx.ProjectDir)

	return nil
}

func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}
