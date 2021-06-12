package sonarqube

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
		Name:      "sonarqube-scan",
		Version:   "0.0.1",
		UsedTools: []string{"sonar-scanner"},
	}
}

// Check evaluates if the action should be executed or not
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

// Execute runs the action
func (action ScanStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	return command.RunOptionalCommand(`sonar-scanner -v`, ctx.Env, ctx.ProjectDir)
}

func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}
