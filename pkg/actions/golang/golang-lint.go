package golang

import (
	"github.com/cidverse/x/pkg/common/command"
	"github.com/rs/zerolog/log"
)

type LintActionStruct struct {
	stage   string
	name    string
	version string
}

// GetStage returns the stage
func (n LintActionStruct) GetStage() string {
	return n.stage
}

// GetName returns the name
func (n LintActionStruct) GetName() string {
	return n.name
}

// GetVersion returns the name
func (n LintActionStruct) GetVersion() string {
	return n.version
}

// SetConfig is used to pass a custom configuration to each action
func (n LintActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (n LintActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectGolangProject(projectDir)
}

// Check if this package can handle the current environment
func (n LintActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	command.RunCommand(`golangci-lint run`, env, projectDir)
}

// LintAction
func LintAction() LintActionStruct {
	entity := LintActionStruct{
		stage: "lint",
		name: "golang-lint",
		version: "0.1.0",
	}

	return entity
}
