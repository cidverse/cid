package owaspdepcheck

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

// Action implementation
type ScanStruct struct {}

// GetDetails returns information about this action
func (action ScanStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "sast",
		Name: "dependencycheck-scan",
		Version: "0.0.1",
		UsedTools: []string{"dependency-check"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action ScanStruct) SetConfig(config string) {
}

// Check if this package can handle the current environment
func (action ScanStruct) Check(projectDir string, env map[string]string) bool {
	return true
}

// Check if this package can handle the current environment
func (action ScanStruct) Execute(projectDir string, env map[string]string, args []string) {
	_ = command.RunOptionalCommand(`dependency-check --noupdate --scan . --enableExperimental --out dist --exclude .git/** --exclude dist/**`, env, projectDir)
}

// BuildAction
func ScanAction() ScanStruct {
	return ScanStruct{}
}
