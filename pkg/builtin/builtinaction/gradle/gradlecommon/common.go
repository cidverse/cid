package gradlecommon

import (
	cidsdk "github.com/cidverse/cid-sdk-go"
)

func GradleTestData(env map[string]string, debug bool) *cidsdk.ModuleActionData {
	env["NCI_COMMIT_REF_TYPE"] = "tag"
	env["NCI_COMMIT_REF_RELEASE"] = "1.0.0"
	return &cidsdk.ModuleActionData{
		ProjectDir: "/my-project",
		Module: cidsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project",
			Discovery:         []cidsdk.ProjectModuleDiscovery{{File: "/my-project/build.gradle.kts"}},
			Name:              "my-module",
			Slug:              "my-module",
			BuildSystem:       string(cidsdk.BuildSystemGradle),
			BuildSystemSyntax: string(cidsdk.BuildSystemSyntaxGradleKotlinDSL),
			Language:          &map[string]string{},
			Submodules:        nil,
		},
		Config: cidsdk.CurrentConfig{
			Debug:       debug,
			Log:         map[string]string{},
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
		},
		Env: env,
	}
}
