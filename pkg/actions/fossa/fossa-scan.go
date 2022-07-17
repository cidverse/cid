package codecov

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

type ScanActionStruct struct{}

// GetDetails retrieves information about the action
func (action ScanActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "fossa-scan",
		Version:   "0.1.0",
		UsedTools: []string{"fossa"},
	}
}

// Check evaluates if the action should be executed or not
func (action ScanActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return len(ctx.Env["FOSSA_API_KEY"]) > 0
}

// Execute runs the action
func (action ScanActionStruct) Execute(ctx *api.ActionExecutionContext, state *api.ActionStateContext) error {
	// env
	ctx.Env["FOSSA_API_KEY"] = ctx.Env["FOSSA_API_KEY"]

	// command
	_ = command.RunOptionalCommand("fossa analyze -o", ctx.Env, ctx.WorkDir)

	return nil
}

func init() {
	api.RegisterBuiltinAction(ScanActionStruct{})
}
