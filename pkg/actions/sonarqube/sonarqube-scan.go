package sonarqube

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"strings"
)

// Action implementation
type ScanStruct struct {}

// GetDetails returns information about this action
func (action ScanStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails {
		Stage: "sast",
		Name: "sonarqube-scan",
		Version: "0.0.1",
		UsedTools: []string{"sonar-scanner"},
	}
}

// Check if this package can handle the current environment
func (action ScanStruct) Check(ctx api.ActionExecutionContext) bool {
	enabled := len(ctx.MachineEnv[SONAR_HOST_URL]) > 0
	if enabled {
		for key, value := range ctx.MachineEnv {
			if strings.HasPrefix(key, SONAR_PREFIX) {
				ctx.Env[key] = value
			}
		}
	}

	return enabled
}

// Check if this package can handle the current environment
func (action ScanStruct) Execute(ctx api.ActionExecutionContext) {
	_ = command.RunOptionalCommand(`sonar-scanner -v`, ctx.Env, ctx.ProjectDir)
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}