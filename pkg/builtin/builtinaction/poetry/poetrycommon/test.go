package poetrycommon

import (
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

func TestModuleData() *actionsdk.ModuleExecutionContextV1Response {
	return &actionsdk.ModuleExecutionContextV1Response{
		Module: &actionsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project",
			Discovery:         []actionsdk.ProjectModuleDiscovery{{File: "/my-project/package.json"}},
			Name:              "my-package",
			Slug:              "my-package",
			BuildSystem:       "python-poetry",
			BuildSystemSyntax: "default",
			Language:          map[string]string{},
			Submodules:        nil,
			Dependencies: []*actionsdk.ProjectDependency{
				{
					Type: "pypi",
					Id:   "pytest",
				},
			},
		},
		Config: &actionsdk.ConfigV1Response{
			Debug:       false,
			Log:         map[string]string{},
			ArtifactDir: ".dist",
			TempDir:     ".tmp",
		},
	}
}
