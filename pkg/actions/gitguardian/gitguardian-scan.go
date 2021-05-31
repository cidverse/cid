package gitguardian

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
		Name: "gitguardian-scan",
		Version: "0.1.0",
		UsedTools: []string{"ggshield"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action ScanStruct) SetConfig(config string) {
}

// Check if this package can handle the current environment
func (action ScanStruct) Check(projectDir string, env map[string]string) bool {
	machineEnv := common.GetMachineEnvironment()
	enabled := len(machineEnv[GITGUARDIAN_API_KEY]) > 0
	if enabled {
		for key, value := range machineEnv {
			if strings.HasPrefix(key, GITGUARDIAN_PREFIX) {
				env[key] = value
			}
		}
	}

	return enabled
}

// Check if this package can handle the current environment
func (action ScanStruct) Execute(projectDir string, env map[string]string, args []string) {
	_ = command.RunOptionalCommand(`ggshield scan path -r -y .`, env, projectDir)
}

// BuildAction
func ScanAction() ScanStruct {
	return ScanStruct{}
}
