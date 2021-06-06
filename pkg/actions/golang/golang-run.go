package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"strings"
)

// Action implementation
type RunActionStruct struct {}

// GetDetails returns all binaries used by this action
func (action RunActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "run",
		Name: "golang-run",
		Version: "0.1.0",
		UsedTools: []string{"go"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action RunActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action RunActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectGolangProject(projectDir)
}

// Check if this package can handle the current environment
func (action RunActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	loadConfig(projectDir)

	_ = command.RunOptionalCommand(`go run . `+strings.Join(args, " "), env, projectDir)
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(RunActionStruct{})
}