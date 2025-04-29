package dotnetcommon

import (
	cidsdk "github.com/cidverse/cid-sdk-go"
)

func ModuleTestData() *cidsdk.ModuleActionData {
	return &cidsdk.ModuleActionData{
		ProjectDir: "/my-project",
		Module: cidsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project",
			Discovery:         []cidsdk.ProjectModuleDiscovery{{File: "/my-project/project.csproj"}},
			Name:              "my-module",
			Slug:              "my-module",
			BuildSystem:       "dotnet",
			BuildSystemSyntax: "default",
			Language:          &map[string]string{},
			Submodules:        nil,
			Files: []string{
				"/my-project/app.go",
			},
		},
		Config: cidsdk.CurrentConfig{
			Log:         map[string]string{},
			ProjectDir:  "/my-project",
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
			Config:      ``,
		},
	}
}
