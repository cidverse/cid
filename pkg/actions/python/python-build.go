package python

import (
	"github.com/PhilippHeuer/cid/pkg/common/command"
	"github.com/rs/zerolog/log"
)

// Action implementation
type BuildActionStruct struct {
	stage   string
	name    string
	version string
}

// GetStage returns the stage
func (n BuildActionStruct) GetStage() string {
	return n.stage
}

// GetName returns the name
func (n BuildActionStruct) GetName() string {
	return n.name
}

// GetVersion returns the name
func (n BuildActionStruct) GetVersion() string {
	return n.version
}

// SetConfig is used to pass a custom configuration to each action
func (n BuildActionStruct) SetConfig(config string) {

}

// Check will evaluate if this action can be executed for the specified project
func (n BuildActionStruct) Check(projectDir string, env []string) bool {
	loadConfig(projectDir)
	return DetectPythonProject(projectDir)
}

// Execute will run the action
func (n BuildActionStruct) Execute(projectDir string, env []string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
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
	entity := BuildActionStruct{
		stage: "build",
		name: "python-build",
		version: "0.1.0",
	}

	return entity
}
