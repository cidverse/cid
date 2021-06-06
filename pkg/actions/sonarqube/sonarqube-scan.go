package sonarqube

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/normalizeci/pkg/common"
	"strings"
)

// Action implementation
type ScanStruct struct {}

// GetDetails returns information about this action
func (action ScanStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "sast",
		Name: "sonarqube-scan",
		Version: "0.0.1",
		UsedTools: []string{"sonar-scanner"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action ScanStruct) SetConfig(config string) {
}

// Check if this package can handle the current environment
func (action ScanStruct) Check(projectDir string, env map[string]string) bool {
	machineEnv := common.GetMachineEnvironment()
	enabled := len(machineEnv[SONAR_HOST_URL]) > 0
	if enabled {
		for key, value := range machineEnv {
			if strings.HasPrefix(key, SONAR_PREFIX) {
				env[key] = value
			}
		}
	}

	return enabled
}

// Check if this package can handle the current environment
func (action ScanStruct) Execute(projectDir string, env map[string]string, args []string) {
	_ = command.RunOptionalCommand(`sonar-scanner -v`, env, projectDir)
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}