package gitguardian

import (
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

type ScanStruct struct{}

// GetDetails retrieves information about the action
func (action ScanStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "gitguardian-scan",
		Version:   "0.1.0",
		UsedTools: []string{"ggshield"},
	}
}

// Check evaluates if the action should be executed or not
func (action ScanStruct) Check(ctx *api.ActionExecutionContext) bool {
	return len(api.GetEnvValue(ctx, GitguardianAPIKey)) > 0
}

// Execute runs the action
func (action ScanStruct) Execute(ctx *api.ActionExecutionContext, state *api.ActionStateContext) error {
	// env
	execEnv := ctx.Env
	execEnv[GitguardianAPIKey] = api.GetEnvValue(ctx, GitguardianAPIKey)

	// GitGuardian env properties
	for key, value := range ctx.MachineEnv {
		if strings.HasPrefix(key, GitguardianPrefix) {
			execEnv[key] = value
		}
	}

	_ = command.RunOptionalCommand(`ggshield scan path -r -y .`, execEnv, ctx.ProjectDir)
	return nil
}

func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}
