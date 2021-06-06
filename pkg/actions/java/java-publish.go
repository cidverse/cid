package java

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/rs/zerolog/log"
	"strings"
)

// Publish
type PublishActionStruct struct {}

// GetDetails returns information about this action
func (action PublishActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails {
		Stage: "publish",
		Name: "java-publish",
		Version: "0.1.0",
		UsedTools: []string{"java"},
	}
}

// Check if this package can handle the current environment
func (action PublishActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectJavaProject(ctx.ProjectDir)
}

// Check if this package can handle the current environment
func (action PublishActionStruct) Execute(ctx api.ActionExecutionContext) {
	// get release version
	releaseVersion := ctx.Env["NCI_COMMIT_REF_RELEASE"]
	// isStableRelease := api.IsVersionStable(releaseVersion)

	// publish
	buildSystem := DetectJavaBuildSystem(ctx.ProjectDir)
	if buildSystem == "gradle-groovy" || buildSystem == "gradle-kotlin" {
		// gradle tasks
		gradleTasks, gradleTasksErr := command.RunSystemCommand(`gradlew`, `tasks --all`, ctx.Env, ctx.ProjectDir)
		if gradleTasksErr != nil {
			log.Warn().Msg("can't list available gradle tasks")
			return
		}

		if strings.Contains(gradleTasks, "publish") {
			command.RunCommand(GradleCommandPrefix+` -Pversion="`+releaseVersion+`" publish --no-daemon --warning-mode=all --console=plain`, ctx.Env, ctx.ProjectDir)
		} else {
			log.Warn().Msg("no supported gradle release plugin found")
		}
	} else if buildSystem == "maven" {
		MavenWrapperSetup(ctx.ProjectDir)

		//
	}
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(PublishActionStruct{})
}