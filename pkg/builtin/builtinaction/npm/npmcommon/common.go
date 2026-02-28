package npmcommon

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
			BuildSystem:       "npm",
			BuildSystemSyntax: "default",
			Language:          map[string]string{},
			Submodules:        nil,
		},
		Config: &actionsdk.ConfigV1Response{
			Log:         map[string]string{},
			ArtifactDir: ".dist",
			TempDir:     ".tmp",
		},
	}
}
