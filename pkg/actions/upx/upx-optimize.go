package upx

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/normalizeci/pkg/common"
)

// Action implementation
type OptimizeActionStruct struct {}

// GetDetails returns information about this action
func (action OptimizeActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "build",
		Name: "upx-optimize",
		Version: "0.1.0",
		UsedTools: []string{"upx"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action OptimizeActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action OptimizeActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)

	machineEnv := common.GetMachineEnvironment()
	return machineEnv["UPX_ENABLED"] == "true"
}

// Check if this package can handle the current environment
func (action OptimizeActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	loadConfig(projectDir)

	command.RunCommand(`upx --lzma `+projectDir+`/`+Config.Paths.Artifact+`/bin/*`, env, projectDir)
}

// OptimizeAction
func OptimizeAction() OptimizeActionStruct {
	return OptimizeActionStruct{}
}
