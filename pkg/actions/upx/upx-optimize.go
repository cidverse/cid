package upx

import (
	"github.com/PhilippHeuer/cid/pkg/common/api"
	"github.com/PhilippHeuer/cid/pkg/common/command"
	"github.com/rs/zerolog/log"
	"os"
)

// Action implementation
type OptimizeActionStruct struct {
	stage   string
	name    string
	version string
}

// GetStage returns the stage
func (n OptimizeActionStruct) GetStage() string {
	return n.stage
}

// GetName returns the name
func (n OptimizeActionStruct) GetName() string {
	return n.name
}

// GetVersion returns the name
func (n OptimizeActionStruct) GetVersion() string {
	return n.version
}

// Check if this package can handle the current environment
func (n OptimizeActionStruct) Check(projectDir string) bool {
	loadConfig(projectDir)

	if _, err := os.Stat(projectDir+`/`+Config.Paths.Artifact+`/bin`); err != nil {
		return false
	}

	return true
}

// Check if this package can handle the current environment
func (n OptimizeActionStruct) Execute(projectDir string, env []string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	env = api.GetEffectiveEnv(env)
	command.RunCommand(`upx --lzma `+projectDir+`/`+Config.Paths.Artifact+`/bin/*`, env)
}

// OptimizeAction
func OptimizeAction() OptimizeActionStruct {
	entity := OptimizeActionStruct{
		stage: "optimize",
		name: "upx-optimize",
		version: "0.1.0",
	}

	return entity
}
