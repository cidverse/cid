package golang

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
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

// SetConfig is used to pass a custom configuration to each action
func (n BuildActionStruct) SetConfig(config string) {
	// parse config
	yamlErr := yaml.Unmarshal([]byte(config), &Config.GoLang)
	if yamlErr != nil {
		log.Fatal().Err(yamlErr).Str("action", n.GetName()).Str("config", config).Msg("failed to parse configuration")
	}
}

// Check if this package can handle the current environment
func (n BuildActionStruct) Check(projectDir string) bool {
	loadConfig(projectDir)
	return DetectGolangProject(projectDir)
}

// Check if this package can handle the current environment
func (n BuildActionStruct) Execute(projectDir string, env []string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	if Config.GoLang.Platform != nil && len(Config.GoLang.Platform) > 0 {
		for _, crossBuild := range Config.GoLang.Platform {
			crossCompile(projectDir, env, crossBuild.Goos, crossBuild.Goarch)
		}
	} else {
		crossCompile(projectDir, env, runtime.GOOS, runtime.GOARCH)
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
