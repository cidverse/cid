package python

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
		Name: "python-build",
		Version: "0.1.0",
		UsedTools: []string{"pipenv", "pip"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action BuildActionStruct) SetConfig(config string) {

}

// Check will evaluate if this action can be executed for the specified project
func (action BuildActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectPythonProject(projectDir)
}

// Execute will run the action
func (action BuildActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	loadConfig(projectDir)

	buildSystem := DetectPythonBuildSystem(projectDir)
	if buildSystem == "requirements.txt" {
		command.RunCommand(`pip install -r requirements.txt`, env, projectDir)
	} else if buildSystem == "pipenv" {
		command.RunCommand(`pipenv install`, env, projectDir)
	} else if buildSystem == "setup.py" {
		command.RunCommand(`pip install ` + projectDir, env, projectDir)
	}
}

// BuildAction
func BuildAction() BuildActionStruct {
	return BuildActionStruct{}
}
