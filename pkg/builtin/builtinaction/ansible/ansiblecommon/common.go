package ansiblecommon

import (
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

func ModuleTestData() *actionsdk.ModuleExecutionContextV1Response {
	return &actionsdk.ModuleExecutionContextV1Response{
		Module: &actionsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project/playbook-a",
			Discovery:         []actionsdk.ProjectModuleDiscovery{{File: "/my-project/playbook-a/playbook.yml"}},
			Name:              "playbook-a",
			Slug:              "playbook-a",
			BuildSystem:       "ansible",
			BuildSystemSyntax: "default",
			Language:          nil,
			Submodules:        nil,
		},
		Config: &actionsdk.ConfigV1Response{
			Debug:       false,
			Log:         map[string]string{},
			ProjectDir:  "/my-project",
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
		},
	}
}
