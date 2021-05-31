package gitleaks

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
)

// Action implementation
type ScanStruct struct {}

// GetDetails returns information about this action
func (action ScanStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "sast",
		Name: "gitleaks-scan",
		Version: "0.1.0",
		UsedTools: []string{"gitleaks"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action ScanStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action ScanStruct) Check(projectDir string, env map[string]string) bool {
	return vcsrepository.GetVCSRepositoryType(projectDir) == "git"
}

// Check if this package can handle the current environment
func (action ScanStruct) Execute(projectDir string, env map[string]string, args []string) {
	_ = command.RunOptionalCommand(`gitleaks --path=. -v --no-git`, env, projectDir)
}

// BuildAction
func GitLeaksScanAction() ScanStruct {
	return ScanStruct{}
}
