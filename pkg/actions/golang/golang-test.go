package golang

import (
	"github.com/PhilippHeuer/cid/pkg/common/command"
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

// Check if this package can handle the current environment
func (n TestActionStruct) Check(projectDir string) bool {
	loadConfig(projectDir)
	return DetectGolangProject(projectDir)
}

// Check if this package can handle the current environment
func (n TestActionStruct) Execute(projectDir string, env []string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	command.RunCommand(`go test -cover ./...`, env)
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
