package owaspdepcheck

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/state"
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

// Execute runs the action
func (action ScanStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	_ = command.RunOptionalCommand(`dependency-check --noupdate --scan . --enableExperimental --out dist --exclude .git/** --exclude `+ctx.Paths.Artifact+`/**`, ctx.Env, ctx.ProjectDir)

	return nil
}

func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}
