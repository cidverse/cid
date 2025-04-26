package uvcommon

import (
	cidsdk "github.com/cidverse/cid-sdk-go"
)

func TestModuleData() *cidsdk.ModuleActionData {
	return &cidsdk.ModuleActionData{
		Module: cidsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project",
			Discovery:         []cidsdk.ProjectModuleDiscovery{{File: "/my-project/package.json"}},
			Name:              "my-package",
			Slug:              "my-package",
			BuildSystem:       "python-uv",
			BuildSystemSyntax: "default",
			Language:          &map[string]string{},
			Submodules:        nil,
			Dependencies: &[]cidsdk.ProjectDependency{
				{
					Type: "pypi",
					Id:   "pytest",
				},
			},
		},
		Config: cidsdk.CurrentConfig{
			Debug:       false,
			Log:         map[string]string{},
			ArtifactDir: ".dist",
			TempDir:     ".tmp",
		},
	}
}
