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
func (action PublishActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "publish",
		Name: "java-publish",
		Version: "0.1.0",
		UsedTools: []string{"java"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action PublishActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action PublishActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectJavaProject(projectDir)
}

// Check if this package can handle the current environment
func (action PublishActionStruct) Execute(projectDirectory string, env map[string]string, args []string) {
	loadConfig(projectDirectory)

	// get release version
	releaseVersion := env["NCI_COMMIT_REF_RELEASE"]
	// isStableRelease := api.IsVersionStable(releaseVersion)

	// publish
	buildSystem := DetectJavaBuildSystem(projectDirectory)
	if buildSystem == "gradle-groovy" || buildSystem == "gradle-kotlin" {
		// gradle tasks
		gradleTasks, gradleTasksErr := command.RunSystemCommand(`gradlew`, `tasks --all`, env, projectDirectory)
		if gradleTasksErr != nil {
			log.Warn().Msg("can't list available gradle tasks")
			return
		}

		if strings.Contains(gradleTasks, "publish") {
			command.RunCommand(GradleCommandPrefix+` -Pversion="`+releaseVersion+`" publish --no-daemon --warning-mode=all --console=plain`, env, projectDirectory)
		} else {
			log.Warn().Msg("no supported gradle release plugin found")
		}
	} else if buildSystem == "maven" {
		MavenWrapperSetup(projectDirectory)

		//
	}
}

// PublishAction
func PublishAction() PublishActionStruct {
	return PublishActionStruct{}
}
