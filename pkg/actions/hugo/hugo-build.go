package hugo

import (
	"github.com/PhilippHeuer/cid/pkg/common/api"
	"github.com/PhilippHeuer/cid/pkg/common/command"
	"github.com/rs/zerolog/log"
)

// Action implementation
type BuildActionStruct struct {
	stage   string
	name    string
	version string
}

// GetStage returns the stage
func (n BuildActionStruct) GetStage() string {
	return n.stage
}

// GetName returns the name
func (n BuildActionStruct) GetName() string {
	return n.name
}

// GetVersion returns the name
func (n BuildActionStruct) GetVersion() string {
	return n.version
}

// Check if this package can handle the current environment
func (n BuildActionStruct) Check(projectDir string) bool {
	return DetectHugoProject(projectDir)
}

// Check if this package can handle the current environment
func (n BuildActionStruct) Execute(projectDir string, env []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	env = api.GetEffectiveEnv(env)
	command.RunCommand(`hugo --minify`, env)
}

// BuildAction
func BuildAction() BuildActionStruct {
	entity := BuildActionStruct{
		stage: "build",
		name: "hugo-build",
		version: "0.1.0",
	}

	return entity
}
