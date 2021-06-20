package golang

import (
	"errors"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/rs/zerolog/log"
	"github.com/shomali11/parallelizer"
	"gopkg.in/yaml.v2"
)

type BuildActionStruct struct{}

// GetDetails retrieves information about the action
func (action BuildActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:            "build",
		Name:             "golang-build",
		Version:          "0.1.0",
		UsedTools:        []string{"go"},
		ToolDependencies: GetToolDependencies(ctx),
	}
}

// Check evaluates if the action should be executed or not
func (action BuildActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return ctx.CurrentModule != nil && ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemGoMod
}

// Execute runs the action
func (action BuildActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	var config Config
	configParseErr := yaml.Unmarshal([]byte(ctx.Config), &config)
	if configParseErr != nil {
		return errors.New("failed to parse action configuration")
	}

	// run build
	if len(config.Platform) > 0 {
		group := parallelizer.NewGroup(parallelizer.WithPoolSize(ctx.Parallelization))
		defer group.Close()

		for _, crossBuild := range config.Platform {
			goos := crossBuild.Goos
			goarch := crossBuild.Goarch

			err := group.Add(func() {
				log.Info().Str("goos", goos).Str("goarch", goarch).Msg("go build")
				CrossCompile(ctx, goos, goarch)
			})
			if err != nil {
				return errors.New("failed to schedule go-build task. Cause: " + err.Error())
			}
		}

		err := group.Wait()
		if err != nil {
			return err
		}
	} else {
		return errors.New("no build configuration present, aborting")
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}
