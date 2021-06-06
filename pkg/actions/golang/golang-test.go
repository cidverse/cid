package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

// Action implementation
type TestActionStruct struct {}

// GetDetails returns information about this action
func (action TestActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "test",
		Name: "golang-test",
		Version: "0.1.0",
		UsedTools: []string{"go"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action TestActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action TestActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectGolangProject(projectDir)
}

// Check if this package can handle the current environment
func (action TestActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	loadConfig(projectDir)

	command.RunCommand(`go test -cover ./...`, env, projectDir)
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(TestActionStruct{})
}