package java

import (
	"errors"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/rs/zerolog/log"
	"strings"
)

type PublishActionStruct struct{}

// GetDetails returns information about this action
func (action PublishActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "publish",
		Name:      "java-publish",
		Version:   "0.1.0",
		UsedTools: []string{"java"},
	}
}

// Check evaluates if the action should be executed or not
func (action PublishActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return ctx.CurrentModule != nil && (ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemGradle || ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemMaven)
}

// Execute runs the action
func (action PublishActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	// get release version
	releaseVersion := ctx.Env["NCI_COMMIT_REF_RELEASE"]
	// isStableRelease := api.IsVersionStable(releaseVersion)

	// publish
	if ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemGradle {
		// gradle tasks
		gradleTasks, gradleTasksErr := command.RunSystemCommand(`gradlew`, `tasks --all`, ctx.Env, ctx.ProjectDir)
		if gradleTasksErr != nil {
			return errors.New("failed to list gradle tasks (gradle tasks --all)")
		}

		if strings.Contains(gradleTasks, "publish") {
			command.RunCommand(GradleCommandPrefix+` -Pversion="`+releaseVersion+`" publish --no-daemon --warning-mode=all --console=plain`, ctx.Env, ctx.ProjectDir)
		} else {
			log.Warn().Msg("no supported gradle release plugin found")
		}
	} else if ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemMaven {
		MavenWrapperSetup(ctx.ProjectDir)

		//
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(PublishActionStruct{})
}
