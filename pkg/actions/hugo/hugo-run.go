package hugo

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

// Action implementation
type RunActionStruct struct {}

// GetDetails returns information about this action
func (action RunActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "run",
		Name: "hugo-run",
		Version: "0.1.0",
		UsedTools: []string{"hugo"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action RunActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action RunActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectHugoProject(projectDir)
}

// Check if this package can handle the current environment
func (action RunActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	loadConfig(projectDir)

	_ = command.RunOptionalCommand(`hugo server --minify --gc --log --verboseLog --baseUrl "/" --watch --source `+projectDir+``, env, projectDir)
}

// BuildAction
func RunAction() RunActionStruct {
	return RunActionStruct{}
}
