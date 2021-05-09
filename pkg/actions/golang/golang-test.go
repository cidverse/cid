package golang

import (
	"github.com/qubid/x/pkg/common/command"
	"github.com/rs/zerolog/log"
)

// Action implementation
type TestActionStruct struct {
	stage   string
	name    string
	version string
}

// GetStage returns the stage
func (n TestActionStruct) GetStage() string {
	return n.stage
}

// GetName returns the name
func (n TestActionStruct) GetName() string {
	return n.name
}

// GetVersion returns the name
func (n TestActionStruct) GetVersion() string {
	return n.version
}

// SetConfig is used to pass a custom configuration to each action
func (n TestActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (n TestActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectGolangProject(projectDir)
}

// Check if this package can handle the current environment
func (n TestActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	command.RunCommand(`go test -cover ./...`, env, projectDir)
}

// BuildAction
func TestAction() TestActionStruct {
	entity := TestActionStruct{
		stage: "test",
		name: "golang-test",
		version: "0.1.0",
	}

	return entity
}
