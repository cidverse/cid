package golang

import (
	"github.com/cidverse/x/pkg/common/command"
	"github.com/rs/zerolog/log"
	"strings"
)

// Action implementation
type RunActionStruct struct {
	stage   string
	name    string
	version string
}

// GetStage returns the stage
func (n RunActionStruct) GetStage() string {
	return n.stage
}

// GetName returns the name
func (n RunActionStruct) GetName() string {
	return n.name
}

// GetVersion returns the name
func (n RunActionStruct) GetVersion() string {
	return n.version
}

// SetConfig is used to pass a custom configuration to each action
func (n RunActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (n RunActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectGolangProject(projectDir)
}

// Check if this package can handle the current environment
func (n RunActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	command.RunCommand(`go run . `+strings.Join(args, " "), env, projectDir)
}

// BuildAction
func RunAction() RunActionStruct {
	entity := RunActionStruct{
		stage: "run",
		name: "golang-run",
		version: "0.1.0",
	}

	return entity
}
