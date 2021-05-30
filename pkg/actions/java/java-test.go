package java

import (
	"github.com/cidverse/x/pkg/common/command"
	"github.com/rs/zerolog/log"
)

// Action implementation
type TestActionStruct struct {
	stage   string
	name    string
	version string
}

// GetStage returns the stage
func (n TestActionStruct) GetStage() string {
	return n.stage
}

// GetName returns the name
func (n TestActionStruct) GetName() string {
	return n.name
}

// GetVersion returns the name
func (n TestActionStruct) GetVersion() string {
	return n.version
}

// SetConfig is used to pass a custom configuration to each action
func (n TestActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (n TestActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectJavaProject(projectDir)
}

// Check if this package can handle the current environment
func (n TestActionStruct) Execute(projectDirectory string, env map[string]string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDirectory)

	// get release version
	releaseVersion := env["NCI_COMMIT_REF_RELEASE"]

	// run test
	buildSystem := DetectJavaBuildSystem(projectDirectory)
	if buildSystem == "gradle-groovy" || buildSystem == "gradle-kotlin" {
		command.RunCommand(GradleCommandPrefix+` -Pversion="`+releaseVersion+`" test --no-daemon --warning-mode=all --console=plain`, env, projectDirectory)
	} else if buildSystem == "maven" {
		MavenWrapperSetup(projectDirectory)

		command.RunCommand(getMavenCommandPrefix(projectDirectory)+" versions:set -DnewVersion="+releaseVersion+"--batch-mode", env, projectDirectory)
		command.RunCommand(getMavenCommandPrefix(projectDirectory)+" test -DskipTests=true --batch-mode", env, projectDirectory)
	}
}

// TestAction
func TestAction() TestActionStruct {
	entity := TestActionStruct{
		stage: "test",
		name: "java-test",
		version: "0.1.0",
	}

	return entity
}
