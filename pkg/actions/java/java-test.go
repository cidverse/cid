package java

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

// Action implementation
type TestActionStruct struct {}

// GetDetails returns information about this action
func (action TestActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails {
		Stage: "test",
		Name: "java-test",
		Version: "0.1.0",
		UsedTools: []string{"java"},
	}
}

// Check if this package can handle the current environment
func (action TestActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectJavaProject(ctx.ProjectDir)
}

// Check if this package can handle the current environment
func (action TestActionStruct) Execute(ctx api.ActionExecutionContext) {
	// get release version
	releaseVersion := ctx.Env["NCI_COMMIT_REF_RELEASE"]

	// run test
	buildSystem := DetectJavaBuildSystem(ctx.ProjectDir)
	if buildSystem == "gradle-groovy" || buildSystem == "gradle-kotlin" {
		command.RunCommand(GradleCommandPrefix+` -Pversion="`+releaseVersion+`" test --no-daemon --warning-mode=all --console=plain`, ctx.Env, ctx.ProjectDir)
	} else if buildSystem == "maven" {
		MavenWrapperSetup(ctx.ProjectDir)

		command.RunCommand(getMavenCommandPrefix(ctx.ProjectDir)+" versions:set -DnewVersion="+releaseVersion+"--batch-mode", ctx.Env, ctx.ProjectDir)
		command.RunCommand(getMavenCommandPrefix(ctx.ProjectDir)+" test -DskipTests=true --batch-mode", ctx.Env, ctx.ProjectDir)
	}
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(TestActionStruct{})
}