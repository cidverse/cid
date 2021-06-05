package python

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

// Action implementation
type CheckActionStruct struct {}

// GetDetails returns information about this action
func (action CheckActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "sast",
		Name: "python-lint",
		Version: "0.1.0",
		UsedTools: []string{"pipenv", "pip"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action CheckActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action CheckActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectPythonProject(projectDir)
}

// Check if this package can handle the current environment
func (action CheckActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	loadConfig(projectDir)

	command.RunCommand(`flake8 .`, env, projectDir)
}

// RunAction
func CheckAction() CheckActionStruct {
	return CheckActionStruct{}
}
