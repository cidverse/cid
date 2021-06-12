package golang

import (
	"errors"
	"github.com/cidverse/cid/pkg/common/api"
	"gopkg.in/yaml.v2"
	"runtime"
)

type BuildActionStruct struct{}

// GetDetails retrieves information about the action
func (action BuildActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:            "build",
		Name:             "golang-build",
		Version:          "0.1.0",
		UsedTools:        []string{"go"},
		ToolDependencies: GetDependencies(ctx.ProjectDir),
	}
}

// Check evaluates if the action should be executed or not
func (action BuildActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectGolangProject(ctx.ProjectDir)
}

// Execute runs the action
func (action BuildActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	var config Config
	configParseErr := yaml.Unmarshal([]byte(ctx.Config), &config)
	if configParseErr != nil {
		return errors.New("failed to parse action configuration")
	}

	// run build
	if config.Platform != nil && len(config.Platform) > 0 {
		for _, crossBuild := range config.Platform {
			CrossCompile(ctx, crossBuild.Goos, crossBuild.Goarch)
		}
	} else {
		CrossCompile(ctx, runtime.GOOS, runtime.GOARCH)
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}
