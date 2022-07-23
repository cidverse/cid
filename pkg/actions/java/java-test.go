package java

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
)

type TestActionStruct struct{}

// GetDetails retrieves information about the action
func (action TestActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "java-test",
		Version:   "0.1.0",
		UsedTools: []string{"java"},
	}
}

// Check evaluates if the action should be executed or not
func (action TestActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return ctx.CurrentModule != nil && (ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemGradle || ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemMaven)
}

// Execute runs the action
func (action TestActionStruct) Execute(ctx *api.ActionExecutionContext, state *api.ActionStateContext) error {
	// get release version
	releaseVersion := ctx.Env["NCI_COMMIT_REF_RELEASE"]

	// run test
	if ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemGradle {
		command.RunCommand(GradleCommandPrefix+` -Pversion="`+releaseVersion+`" check --no-daemon --warning-mode=all --console=plain`, ctx.Env, ctx.ProjectDir)

		// collect jacoco reports from all modules
		processJacocoFile(ctx, ctx.CurrentModule, "build/reports/jacoco/test/jacocoTestReport.xml")
		for _, submodule := range ctx.CurrentModule.Submodules {
			processJacocoFile(ctx, submodule, "build/reports/jacoco/test/jacocoTestReport.xml")
		}
	} else if ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemMaven {
		MavenWrapperSetup(ctx.ProjectDir)

		command.RunCommand(getMavenCommandPrefix(ctx.ProjectDir)+" versions:set -DnewVersion="+releaseVersion+" --batch-mode", ctx.Env, ctx.ProjectDir)
		command.RunCommand(getMavenCommandPrefix(ctx.ProjectDir)+" test -DskipTests=true --batch-mode", ctx.Env, ctx.ProjectDir)

		// collect jacoco reports from all modules
		processJacocoFile(ctx, ctx.CurrentModule, "target/site/jacoco/jacoco.xml")
		for _, submodule := range ctx.CurrentModule.Submodules {
			processJacocoFile(ctx, submodule, "target/site/jacoco/jacoco.xml")
		}
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(TestActionStruct{})
}
