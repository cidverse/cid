package hugo

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
		Name: "hugo-build",
		Version: "0.1.0",
		UsedTools: []string{"hugo"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action BuildActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action BuildActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectHugoProject(projectDir)
}

// Check if this package can handle the current environment
func (action BuildActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	loadConfig(projectDir)

	command.RunCommand(`hugo --minify --gc --log --verboseLog --source `+projectDir+` --destination `+ projectDir+`/`+Config.Paths.Artifact, env, projectDir)
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}