package sonarqube

import (
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"path/filepath"
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
	scanArgs = append(scanArgs, `-D sonar.tests=.`)

	// module specific parameters
	var sourceInclusion []string
	var sourceExclusions []string
	var testInclusion []string
	var testExclusions []string
	for _, module := range ctx.Modules {
		if module.BuildSystem == analyzerapi.BuildSystemGradle || module.BuildSystem == analyzerapi.BuildSystemMaven {
			sourceInclusion = append(sourceInclusion, "**/src/main/java/**", "**/src/main/kotlin/**")
			testInclusion = append(testInclusion, "**/src/test/java/**", "**/src/test/kotlin/**")
			scanArgs = append(scanArgs, `-D sonar.coverage.jacoco.xmlReportPaths=`+filepath.Join(ctx.Paths.Artifact, "**", "test", "jacoco.xml"))
			scanArgs = append(scanArgs, `-D sonar.java.binaries=.`)
			scanArgs = append(scanArgs, `-D sonar.java.test.binaries=.`)
		} else if module.BuildSystem == analyzerapi.BuildSystemGoMod {
			sourceExclusions = append(sourceExclusions, "**/*_test.go", "**/vendor/**", "**/testdata/*")
			testInclusion = append(testInclusion, "**/*_test.go")
			testExclusions = append(testExclusions, "**/vendor/**")
			scanArgs = append(scanArgs, `-D sonar.go.coverage.reportPaths=`+filepath.Join(ctx.Paths.ArtifactModule(module.Slug, "test"), "coverage.out"))
			scanArgs = append(scanArgs, `-D sonar.go.tests.reportPaths=`+filepath.Join(ctx.Paths.ArtifactModule(module.Slug, "test"), "coverage.json"))
		}
	}
	scanArgs = append(scanArgs, `-D sonar.inclusions=`+strings.Join(sourceInclusion, ","))
	scanArgs = append(scanArgs, `-D sonar.exclusions=`+strings.Join(sourceExclusions, ","))
	scanArgs = append(scanArgs, `-D sonar.test.inclusions=`+strings.Join(testInclusion, ","))
	scanArgs = append(scanArgs, `-D sonar.test.exclusions=`+strings.Join(testExclusions, ","))

	return command.RunOptionalCommand(strings.Join(scanArgs, " "), ctx.Env, ctx.ProjectDir)
}

func init() {
	api.RegisterBuiltinAction(ScanStruct{})
}
