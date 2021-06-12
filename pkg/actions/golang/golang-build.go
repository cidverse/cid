package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"runtime"
)

type BuildActionStruct struct{}

func (action BuildActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:            "build",
		Name:             "golang-build",
		Version:          "0.1.0",
		UsedTools:        []string{"go"},
		ToolDependencies: GetDependencies(ctx.ProjectDir),
	}
}

func (action BuildActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectGolangProject(ctx.ProjectDir)
}

func (action BuildActionStruct) Execute(ctx api.ActionExecutionContext) {
	var config Config
	configParseErr := yaml.Unmarshal([]byte(ctx.Config), &config)
	if configParseErr != nil {
		log.Error().Err(configParseErr).Str("action", "golang-build").Msg("failed to parse action configuration")
		return
	}

	// run build
	if config.Platform != nil && len(config.Platform) > 0 {
		for _, crossBuild := range config.Platform {
			CrossCompile(ctx, crossBuild.Goos, crossBuild.Goarch)
		}
	} else {
		CrossCompile(ctx, runtime.GOOS, runtime.GOARCH)
	}
}

func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}
