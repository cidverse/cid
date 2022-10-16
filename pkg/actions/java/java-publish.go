package java

import (
	"errors"
	"github.com/cidverse/cid/pkg/core/state"
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/rs/zerolog/log"
)

type PublishActionStruct struct{}

// GetDetails returns information about this action
func (action PublishActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "java-publish",
		Version:   "0.1.0",
		UsedTools: []string{"java"},
	}
}

// Check evaluates if the action should be executed or not
func (action PublishActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return true
}

// Execute runs the action
func (action PublishActionStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	// release env
	releaseEnv := ctx.Env
	// - gpg signing
	releaseEnv["ORG_GRADLE_PROJECT_signingKeyId"] = api.GetEnvValue(ctx, "GPG_SIGN_KEYID")
	releaseEnv["ORG_GRADLE_PROJECT_signingKey"] = api.GetEnvValue(ctx, "GPG_SIGN_PRIVATEKEY")
	releaseEnv["ORG_GRADLE_PROJECT_signingPassword"] = api.GetEnvValue(ctx, "GPG_SIGN_PASSWORD")
	// - repo
	releaseEnv["MAVEN_REPO_URL"] = api.GetEnvValue(ctx, "MAVEN_REPO_URL")
	releaseEnv["MAVEN_REPO_USERNAME"] = api.GetEnvValue(ctx, "MAVEN_REPO_USERNAME")
	releaseEnv["MAVEN_REPO_PASSWORD"] = api.GetEnvValue(ctx, "MAVEN_REPO_PASSWORD")

	// publish
	if ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemGradle {
		// gradle tasks
		gradleTasks, gradleTasksErr := command.RunCommandAndGetOutput(GradleCommandPrefix+` tasks --all`, ctx.Env, ctx.ProjectDir)
		if gradleTasksErr != nil {
			return errors.New("failed to list gradle tasks (gradle tasks --all)")
		}

		if strings.Contains(gradleTasks, "publish") {
			command.RunCommand(api.ReplacePlaceholders(GradleCommandPrefix+` -Pversion="{NCI_COMMIT_REF_RELEASE}" publish --no-daemon --warning-mode=all --console=plain --stacktrace`, ctx.Env), releaseEnv, ctx.ProjectDir)
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
