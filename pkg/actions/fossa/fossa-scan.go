package fossa

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/state"
	"strings"
)

type ScanActionStruct struct{}

// GetDetails retrieves information about the action
func (action ScanActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "fossa-source-scan",
		Version:   "0.1.0",
		UsedTools: []string{"fossa"},
	}
}

// Check evaluates if the action should be executed or not
func (action ScanActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return len(ctx.Env["FOSSA_API_KEY"]) > 0
}

// Execute runs the action
func (action ScanActionStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	// run scan
	var scanArgs []string
	scanArgs = append(scanArgs, `fossa analyze`)
	_ = command.RunOptionalCommand(strings.Join(scanArgs, " "), ctx.Env, ctx.ProjectDir)

	return nil
}

func init() {
	api.RegisterBuiltinAction(ScanActionStruct{})
}
