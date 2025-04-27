package gocommon

import (
	cidsdk "github.com/cidverse/cid-sdk-go"
)

func ModuleTestData() *cidsdk.ModuleActionData {
	return &cidsdk.ModuleActionData{
		ProjectDir: "/my-project",
		Module: cidsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project",
			Discovery:         []cidsdk.ProjectModuleDiscovery{{File: "/my-project/go.mod"}},
			Name:              "github.com/cidverse/my-project",
			Slug:              "github-com-cidverse-my-project",
			BuildSystem:       "gomod",
			BuildSystemSyntax: "default",
			Language:          &map[string]string{"go": "1.19.0"},
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
			Config:      `{"platform":[{"goos":"linux","goarch":"amd64"}]}`,
		},
	}
}
