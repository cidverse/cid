package java

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

// Action implementation
type TestActionStruct struct {}

// GetDetails returns information about this action
func (action TestActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "test",
		Name: "java-test",
		Version: "0.1.0",
		UsedTools: []string{"java"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action TestActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action TestActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectJavaProject(projectDir)
}

// Check if this package can handle the current environment
func (action TestActionStruct) Execute(projectDirectory string, env map[string]string, args []string) {
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
	return TestActionStruct{}
}
