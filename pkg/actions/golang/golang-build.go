package golang

import (
	"errors"
	"github.com/cidverse/cid/pkg/core/state"
	"os"
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/rs/zerolog/log"
	"github.com/shomali11/parallelizer"
	"gopkg.in/yaml.v2"
)

type BuildActionStruct struct{}

// GetDetails retrieves information about the action
func (action BuildActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:             "golang-build",
		Version:          "0.1.0",
		UsedTools:        []string{"go"},
		ToolDependencies: GetToolDependencies(ctx),
	}
}

// Check evaluates if the action should be executed or not
func (action BuildActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return true
}

// Execute runs the action
func (action BuildActionStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	var config Config
	configParseErr := yaml.Unmarshal([]byte(ctx.Config), &config)
	if configParseErr != nil {
		return errors.New("failed to parse action configuration")
	}

	// run build
	if len(config.Platform) > 0 {
		group := parallelizer.NewGroup(parallelizer.WithPoolSize(ctx.Parallelization))
		defer group.Close()

		// install and build only works for projects that generate binaries, not for library modules
		if !hasGoFilesInDirectory(ctx.CurrentModule.FilesByExtension["go"], ctx.CurrentModule.RootDirectory) {
			log.Info().Msg("no go files found in main directory, assuming go library, skipping local installation and binary build.")
			return nil
		}

		// install locally
		if !strings.EqualFold(ctx.Env["CI"], "true") {
			err := group.Add(func() {
				log.Info().Msg("go install")
				command.RunCommand(api.ReplacePlaceholders(`go install -buildvcs=false -ldflags "`+GetLdFlags(config)+`-X main.Version={NCI_COMMIT_REF_RELEASE} -X main.RepositoryStatus={NCI_REPOSITORY_STATUS} -X main.CommitHash={NCI_COMMIT_SHA} -X main.BuildAt={NOW_RFC3339}" .`, ctx.Env), ctx.Env, ctx.CurrentModule.Directory)
			})
			if err != nil {
				return errors.New("failed to schedule go-install task. Cause: " + err.Error())
			}
		}

		// compile
		for _, crossBuild := range config.Platform {
			goos := crossBuild.Goos
			goarch := crossBuild.Goarch

			err := group.Add(func() {
				log.Info().Str("goos", goos).Str("goarch", goarch).Msg("go build")
				CrossCompile(ctx, config, goos, goarch)
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

// hasGoFilesInDirectory checks of the file list contains go files in the root directory
func hasGoFilesInDirectory(files []string, directory string) bool {
	for _, file := range files {
		fileRelative := strings.TrimPrefix(file, directory+string(os.PathSeparator))
		if !strings.ContainsRune(fileRelative, os.PathSeparator) {
			return true
		}
	}
	return false
}
