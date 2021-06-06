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
func (action BuildActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails {
		Stage: "build",
		Name: "golang-build",
		Version: "0.1.0",
		UsedTools: []string{"go"},
		ToolDependencies: GetDependencies(ctx.ProjectDir),
	}
}

// Check if this package can handle the current environment
func (action BuildActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectGolangProject(ctx.ProjectDir)
}

// Check if this package can handle the current environment
func (action BuildActionStruct) Execute(ctx api.ActionExecutionContext) {
	// load config
	yamlErr := yaml.Unmarshal([]byte(ctx.Config), &Config.GoLang)
	if yamlErr != nil {
		log.Fatal().Err(yamlErr).Str("config", ctx.Config).Msg("failed to parse configuration")
	}

	// run build
	if Config.GoLang.Platform != nil && len(Config.GoLang.Platform) > 0 {
		for _, crossBuild := range Config.GoLang.Platform {
			CrossCompile(ctx, crossBuild.Goos, crossBuild.Goarch)
		}
	} else {
		CrossCompile(ctx, runtime.GOOS, runtime.GOARCH)
	}
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}