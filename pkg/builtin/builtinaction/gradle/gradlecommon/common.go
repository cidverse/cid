package gradlecommon

import (
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/repoanalyzer/analyzerapi"
)

func GradleTestData(env map[string]string, debug bool) *actionsdk.ModuleExecutionContextV1Response {
	env["NCI_COMMIT_REF_TYPE"] = "tag"
	env["NCI_COMMIT_REF_RELEASE"] = "1.0.0"
	return &actionsdk.ModuleExecutionContextV1Response{
		ProjectDir: "/my-project",
		Module: &actionsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project",
			Discovery:         []actionsdk.ProjectModuleDiscovery{{File: "/my-project/build.gradle.kts"}},
			Name:              "my-module",
			Slug:              "my-module",
			BuildSystem:       string(analyzerapi.BuildSystemGradle),
			BuildSystemSyntax: string(analyzerapi.BuildSystemSyntaxGradleKotlinDSL),
			Language:          map[string]string{},
			Submodules:        nil,
		},
		Config: &actionsdk.ConfigV1Response{
			Debug:       debug,
			Log:         map[string]string{},
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
		},
		Env: env,
	}
}
