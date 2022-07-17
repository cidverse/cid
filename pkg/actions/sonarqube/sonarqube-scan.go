package sonarqube

import (
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/normalizeci/pkg/ncispec"
)

type ScanStruct struct{}

// GetDetails retrieves information about the action
func (action ScanStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "sonarqube-scan",
		Version:   "0.0.1",
		UsedTools: []string{"sonar-scanner"},
	}
}

// Check evaluates if the action should be executed or not
func (action ScanStruct) Check(ctx *api.ActionExecutionContext) bool {
	return len(ctx.MachineEnv[SonarHostURL]) > 0
}

// Execute runs the action
func (action ScanStruct) Execute(ctx *api.ActionExecutionContext, state *api.ActionStateContext) error {
	// env
	for key, value := range ctx.MachineEnv {
		if strings.HasPrefix(key, SonarPrefix) {
			ctx.Env[key] = value
		}
	}

	return command.RunOptionalCommand(`sonar-scanner -v`, ctx.Env, ctx.ProjectDir)
}

func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}
