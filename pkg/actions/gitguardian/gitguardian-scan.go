package gitguardian

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"strings"
)

type ScanStruct struct{}

// GetDetails retrieves information about the action
func (action ScanStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "sast",
		Name:      "gitguardian-scan",
		Version:   "0.1.0",
		UsedTools: []string{"ggshield"},
	}
}

// Check evaluates if the action should be executed or not
func (action ScanStruct) Check(ctx api.ActionExecutionContext) bool {
	return len(api.GetEnvValue(ctx, GITGUARDIAN_API_KEY)) > 0
}

// Execute runs the action
func (action ScanStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	// env
	execEnv := ctx.Env
	execEnv[GITGUARDIAN_API_KEY] = api.GetEnvValue(ctx, GITGUARDIAN_API_KEY)
	// gitguardian env  properties
	for key, value := range ctx.MachineEnv {
		if strings.HasPrefix(key, GITGUARDIAN_PREFIX) {
			execEnv[key] = value
		}
	}

	_ = command.RunOptionalCommand(`ggshield scan path -r -y .`, execEnv, ctx.ProjectDir)
	return nil
}

func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}
