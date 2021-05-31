package upx

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

// Action implementation
type BuildActionStruct struct {}

// GetDetails returns information about this action
func (action BuildActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "build",
		Name: "upx-optimize",
		Version: "0.1.0",
		UsedTools: []string{"upx"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action BuildActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action BuildActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)

	return false
}

// Check if this package can handle the current environment
func (action BuildActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	loadConfig(projectDir)

	command.RunCommand(`upx --lzma `+projectDir+`/`+Config.Paths.Artifact+`/bin/*`, env, projectDir)
}

// BuildAction
func BuildAction() BuildActionStruct {
	return BuildActionStruct{}
}
