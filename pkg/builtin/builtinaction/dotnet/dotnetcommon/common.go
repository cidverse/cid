package dotnetcommon

import (
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

func ModuleTestData() *actionsdk.ModuleExecutionContextV1Response {
	return &actionsdk.ModuleExecutionContextV1Response{
		ProjectDir: "/my-project",
		Module: &actionsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project",
			Discovery:         []actionsdk.ProjectModuleDiscovery{{File: "/my-project/project.csproj"}},
			Name:              "my-module",
			Slug:              "my-module",
			BuildSystem:       "dotnet",
			BuildSystemSyntax: "default",
			Language:          map[string]string{},
			Submodules:        nil,
			Files: []string{
				"/my-project/app.go",
			},
		},
		Config: &actionsdk.ConfigV1Response{
			Log:         map[string]string{},
			ProjectDir:  "/my-project",
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
			Config:      ``,
		},
	}
}
