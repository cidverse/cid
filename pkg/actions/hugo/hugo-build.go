package hugo

import (
	"github.com/qubid/x/pkg/common/command"
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

// SetConfig is used to pass a custom configuration to each action
func (n BuildActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (n BuildActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectHugoProject(projectDir)
}

// Check if this package can handle the current environment
func (n BuildActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	command.RunCommand(`hugo --minify --gc --log --verboseLog --source `+projectDir+` --destination `+ projectDir+`/`+Config.Paths.Artifact, env, projectDir)
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
