package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

type LintActionStruct struct {}

// GetDetails returns information about this action
func (action LintActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "sast",
		Name: "golang-lint",
		Version: "0.1.0",
		UsedTools: []string{"golangci-lint"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action LintActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action LintActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectGolangProject(projectDir)
}

// Check if this package can handle the current environment
func (action LintActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	loadConfig(projectDir)

	command.RunCommand(`golangci-lint run`, env, projectDir)
}

// LintAction
func LintAction() LintActionStruct {
	return LintActionStruct{}
}
