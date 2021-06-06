package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"runtime"
)

// Action implementation
type BuildActionStruct struct {}

// GetDetails returns information about this action
func (action BuildActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "build",
		Name: "golang-build",
		Version: "0.1.0",
		UsedTools: []string{"go"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action BuildActionStruct) SetConfig(config string) {
	// parse config
	yamlErr := yaml.Unmarshal([]byte(config), &Config.GoLang)
	if yamlErr != nil {
		log.Fatal().Err(yamlErr).Str("config", config).Msg("failed to parse configuration")
	}
}

// Check if this package can handle the current environment
func (action BuildActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectGolangProject(projectDir)
}

// Check if this package can handle the current environment
func (action BuildActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	loadConfig(projectDir)

	if Config.GoLang.Platform != nil && len(Config.GoLang.Platform) > 0 {
		for _, crossBuild := range Config.GoLang.Platform {
			CrossCompile(projectDir, env, crossBuild.Goos, crossBuild.Goarch)
		}
	} else {
		CrossCompile(projectDir, env, runtime.GOOS, runtime.GOARCH)
	}
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}