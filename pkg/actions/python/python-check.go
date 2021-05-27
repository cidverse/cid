package python

import (
	"github.com/cidverse/x/pkg/common/command"
	"github.com/rs/zerolog/log"
)

// Action implementation
type CheckActionStruct struct {
	stage   string
	name    string
	version string
}

// GetStage returns the stage
func (n CheckActionStruct) GetStage() string {
	return n.stage
}

// GetName returns the name
func (n CheckActionStruct) GetName() string {
	return n.name
}

// GetVersion returns the name
func (n CheckActionStruct) GetVersion() string {
	return n.version
}

// SetConfig is used to pass a custom configuration to each action
func (n CheckActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (n CheckActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectPythonProject(projectDir)
}

// Check if this package can handle the current environment
func (n CheckActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	command.RunCommand(`flake8 .`, env, projectDir)
}

// RunAction
func CheckAction() CheckActionStruct {
	entity := CheckActionStruct{
		stage: "check",
		name: "python-check",
		version: "0.1.0",
	}

	return entity
}
