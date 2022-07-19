package sonarqube

import (
	"github.com/cidverse/cid/pkg/actions/java"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"strings"

	"github.com/cidverse/cid/pkg/common/protectoutput"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/normalizeci/pkg/ncispec"
)

type ScanStruct struct{}

// GetDetails retrieves information about the action
func (action ScanStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "sonarqube-scan",
		Version:   "0.0.1",
		UsedTools: []string{"sonar-scanner"},
	}
}

// Check evaluates if the action should be executed or not
func (action ScanStruct) Check(ctx *api.ActionExecutionContext) bool {
	return true
}

// Execute runs the action
func (action ScanStruct) Execute(ctx *api.ActionExecutionContext, state *api.ActionStateContext) error {
	// protect token
	protectoutput.ProtectPhrase(ctx.Env[SonarToken])

	// default to cloud host
	if ctx.Env[SonarHostURL] == "" {
		ctx.Env[SonarHostURL] = SonarCloudURL
	}
	if ctx.Env[SonarProjectKey] == "" {
		ctx.Env[SonarProjectKey] = ctx.Env[ncispec.NCI_PROJECT_ID]
	}
	if ctx.Env[SonarDefaultBranch] == "" {
		ctx.Env[SonarDefaultBranch] = "develop"
	}

	// ensure that the default branch is configured correctly
	prepareProject(ctx.Env[SonarHostURL], ctx.Env[SonarToken], ctx.Env[SonarOrganization], ctx.Env[SonarProjectKey], ctx.Env[ncispec.NCI_PROJECT_NAME], ctx.Env[ncispec.NCI_PROJECT_DESCRIPTION], ctx.Env[SonarDefaultBranch])

	// run scan
	var scanArgs []string
	scanArgs = append(scanArgs, `sonar-scanner`)
	scanArgs = append(scanArgs, `-D sonar.host.url=`+ctx.Env[SonarHostURL])
	scanArgs = append(scanArgs, `-D sonar.login=`+ctx.Env[SonarToken])
	if ctx.Env[SonarOrganization] != "" {
		scanArgs = append(scanArgs, `-D sonar.organization=`+ctx.Env[SonarOrganization])
	}
	scanArgs = append(scanArgs, `-D sonar.projectKey=`+ctx.Env[SonarProjectKey])
	scanArgs = append(scanArgs, `-D sonar.projectName=`+ctx.Env[ncispec.NCI_PROJECT_NAME])
	scanArgs = append(scanArgs, `-D sonar.branch.name=`+ctx.Env[ncispec.NCI_COMMIT_REF_SLUG])
	scanArgs = append(scanArgs, `-D sonar.sources=.`)

	// analysis tasks that require compiled code
	scanArgs = append(scanArgs, `-D sonar.java.binaries=.`)
	for _, module := range ctx.Modules {
		if module.BuildSystem == analyzerapi.BuildSystemGradle || module.BuildSystem == analyzerapi.BuildSystemMaven {
			java.BuildJavaProject(ctx, state, module)
		}
	}

	return command.RunOptionalCommand(strings.Join(scanArgs, " "), ctx.Env, ctx.ProjectDir)
}

func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}
