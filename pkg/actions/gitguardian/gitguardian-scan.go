package gitguardian

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"strings"
)

// Action implementation
type ScanStruct struct{}

// GetDetails returns information about this action
func (action ScanStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "sast",
		Name:      "gitguardian-scan",
		Version:   "0.1.0",
		UsedTools: []string{"ggshield"},
	}
}

// Check if this package can handle the current environment
func (action ScanStruct) Check(ctx api.ActionExecutionContext) bool {
	enabled := len(ctx.MachineEnv[GITGUARDIAN_API_KEY]) > 0
	if enabled {
		for key, value := range ctx.MachineEnv {
			if strings.HasPrefix(key, GITGUARDIAN_PREFIX) {
				ctx.Env[key] = value
			}
		}
	}

	return enabled
}

// Check if this package can handle the current environment
func (action ScanStruct) Execute(ctx api.ActionExecutionContext) {
	_ = command.RunOptionalCommand(`ggshield scan path -r -y .`, ctx.Env, ctx.ProjectDir)
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}
