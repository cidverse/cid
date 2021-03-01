package golang

import (
	"github.com/rs/zerolog/log"
	"runtime"
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
	return DetectGolangProject(projectDir)
}

// Check if this package can handle the current environment
func (n BuildActionStruct) Execute(projectDir string, env []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	if GoConfig.GoLang.Platform != nil && len(GoConfig.GoLang.Platform) > 0 {
		for _, crossBuild := range GoConfig.GoLang.Platform {
			crossCompile(env, crossBuild.Goos, crossBuild.Goarch)
		}
	} else {
		crossCompile(env, runtime.GOOS, runtime.GOARCH)
	}
}

// BuildAction
func BuildAction() BuildActionStruct {
	entity := BuildActionStruct{
		stage: "build",
		name: "golang-build",
		version: "0.1.0",
	}

	return entity
}
